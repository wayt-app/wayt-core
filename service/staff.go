package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wayt/wayt-core/model"
	"github.com/wayt/wayt-core/repository"
	"github.com/wayt/wayt-core/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

func generateStaffToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type StaffService interface {
	Create(ownerID, branchID uint, name, emailAddr, password string) (*model.Staff, error)
	Login(emailAddr, password string) (string, error)
	List(ownerID uint) ([]model.Staff, error)
	FindByID(id uint) (*model.Staff, error)
	Update(id, ownerID uint, name string, isActive bool) (*model.Staff, error)
	Delete(id, ownerID uint) error
	ChangePassword(id uint, currentPassword, newPassword string) error
	ForgotPassword(emailAddr string) error
	ResetPasswordWithToken(token, newPassword string) error
	Logout(id uint) error
}

type staffService struct {
	repo           repository.StaffRepository
	branchRepo     repository.BranchRepository
	restaurantRepo repository.RestaurantRepository
	emailSvc       email.Sender
	jwtSecret      []byte
}

func NewStaffService(
	repo repository.StaffRepository,
	branchRepo repository.BranchRepository,
	restaurantRepo repository.RestaurantRepository,
	emailSvc email.Sender,
	jwtSecret string,
) StaffService {
	return &staffService{
		repo:           repo,
		branchRepo:     branchRepo,
		restaurantRepo: restaurantRepo,
		emailSvc:       emailSvc,
		jwtSecret:      []byte(jwtSecret),
	}
}

func (s *staffService) Create(ownerID, branchID uint, name, emailAddr, password string) (*model.Staff, error) {
	if name == "" || emailAddr == "" || password == "" {
		return nil, errors.New("nama, email, dan password wajib diisi")
	}

	// Validate branch belongs to owner's restaurant
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	restaurant, err := s.restaurantRepo.FindByID(branch.RestaurantID)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	if restaurant.BusinessOwnerID == nil || *restaurant.BusinessOwnerID != ownerID {
		return nil, errors.New("cabang tidak berada di restoran Anda")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	st := &model.Staff{
		BusinessOwnerID: ownerID,
		BranchID:        branchID,
		Name:            name,
		Email:           emailAddr,
		Password:        string(hashed),
		IsActive:        true,
	}
	if err := s.repo.Create(st); err != nil {
		return nil, errors.New("email sudah digunakan")
	}
	return st, nil
}

func (s *staffService) Login(emailAddr, password string) (string, error) {
	st, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		return "", errors.New("email atau password salah")
	}
	if !st.IsActive {
		return "", errors.New("akun staff tidak aktif")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(st.Password), []byte(password)); err != nil {
		return "", errors.New("email atau password salah")
	}

	// Get restaurant_id via branch
	var restaurantID uint
	branch, err := s.branchRepo.FindByID(st.BranchID)
	if err == nil {
		restaurantID = branch.RestaurantID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":           st.ID,
		"name":          st.Name,
		"email":         st.Email,
		"type":          "staff",
		"branch_id":     st.BranchID,
		"restaurant_id": restaurantID,
		"token_version": st.TokenVersion,
		"exp":           time.Now().Add(12 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}

func (s *staffService) List(ownerID uint) ([]model.Staff, error) {
	return s.repo.FindByOwnerID(ownerID)
}

func (s *staffService) FindByID(id uint) (*model.Staff, error) {
	st, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("staff tidak ditemukan")
	}
	return st, nil
}

func (s *staffService) Update(id, ownerID uint, name string, isActive bool) (*model.Staff, error) {
	st, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("staff tidak ditemukan")
	}
	if st.BusinessOwnerID != ownerID {
		return nil, errors.New("tidak diizinkan mengubah staff ini")
	}
	if name != "" {
		st.Name = name
	}
	st.IsActive = isActive
	if err := s.repo.Update(st); err != nil {
		return nil, fmt.Errorf("gagal update staff: %w", err)
	}
	return st, nil
}

func (s *staffService) Delete(id, ownerID uint) error {
	st, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("staff tidak ditemukan")
	}
	if st.BusinessOwnerID != ownerID {
		return errors.New("tidak diizinkan menghapus staff ini")
	}
	return s.repo.Delete(id)
}

func (s *staffService) ChangePassword(id uint, currentPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	st, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("staff tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(st.Password), []byte(currentPassword)); err != nil {
		return errors.New("password saat ini salah")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, string(hashed))
}

func (s *staffService) ForgotPassword(emailAddr string) error {
	st, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		return nil // don't leak if email exists
	}
	token, err := generateStaffToken()
	if err != nil {
		return errors.New("gagal membuat token")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	if err := s.repo.SetResetToken(st.ID, token, expiresAt); err != nil {
		return errors.New("gagal menyimpan token")
	}
	html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Anda menerima permintaan reset password untuk akun Staff Wayt Business Anda.</p>
<p>Token reset password Anda (berlaku 15 menit):</p>
<p style="font-size:24px;font-weight:bold;letter-spacing:4px;color:#7c3aed">%s</p>
<p>Masukkan token ini di halaman reset password staff. Jika Anda tidak meminta reset password, abaikan email ini.</p>
`, st.Name, token)
	if err := s.emailSvc.Send(emailAddr, "Reset Password Staff — Wayt Business", html); err != nil {
		log.Printf("[EMAIL ERROR] reset password staff %s: %v", emailAddr, err)
	}
	return nil
}

func (s *staffService) ResetPasswordWithToken(token, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	st, err := s.repo.FindByResetToken(token)
	if err != nil {
		return errors.New("token tidak valid atau sudah kadaluarsa")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(st.ID, string(hashed)); err != nil {
		return err
	}
	return s.repo.ClearResetToken(st.ID)
}

func (s *staffService) Logout(id uint) error {
	return s.repo.IncrementTokenVersion(id)
}
