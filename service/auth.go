package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wayt/wayt-core/model"
	"github.com/wayt/wayt-core/repository"
	"github.com/wayt/wayt-core/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username, password string) (string, error)
	SeedSuperAdmin(username, password string) error
	ListAdmins() ([]model.AdminUser, error)
	CreateAdmin(username, password string, role model.AdminRole, restaurantID *uint) (*model.AdminUser, error)
	UpdateAdmin(id uint, username string, role model.AdminRole, password string, restaurantID *uint) (*model.AdminUser, error)
	DeleteAdmin(id uint, requesterID uint) error
	ForgotPassword(username string) (string, error)
	ResetPasswordWithToken(token, newPassword string) error
	Logout(id uint) error
}

type authService struct {
	repo      repository.AdminUserRepository
	jwtSecret []byte
	emailSvc  email.Sender
}

func NewAuthService(repo repository.AdminUserRepository, jwtSecret string, emailSvc email.Sender) AuthService {
	return &authService{repo: repo, jwtSecret: []byte(jwtSecret), emailSvc: emailSvc}
}

func (s *authService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", errors.New("username atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("username atau password salah")
	}
	claims := jwt.MapClaims{
		"sub":           user.ID,
		"username":      user.Username,
		"role":          string(user.Role),
		"token_version": user.TokenVersion,
		"exp":           time.Now().Add(8 * time.Hour).Unix(),
	}
	if user.RestaurantID != nil {
		claims["restaurant_id"] = *user.RestaurantID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}

func (s *authService) SeedSuperAdmin(username, password string) error {
	exists, err := s.repo.ExistsAny()
	if err != nil || exists {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(&model.AdminUser{
		Username: username,
		Role:     model.RoleSuperAdmin,
		Password: string(hashed),
	})
}

func (s *authService) ListAdmins() ([]model.AdminUser, error) {
	return s.repo.FindAll()
}

func (s *authService) CreateAdmin(username, password string, role model.AdminRole, restaurantID *uint) (*model.AdminUser, error) {
	if username == "" || password == "" {
		return nil, errors.New("username dan password wajib diisi")
	}
	if role != model.RoleSuperAdmin && role != model.RoleAdmin {
		role = model.RoleAdmin
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.AdminUser{
		Username:     username,
		Role:         role,
		Password:     string(hashed),
		RestaurantID: restaurantID,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("username sudah digunakan")
	}
	return user, nil
}

func (s *authService) UpdateAdmin(id uint, username string, role model.AdminRole, password string, restaurantID *uint) (*model.AdminUser, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("admin tidak ditemukan")
	}
	if username != "" {
		user.Username = username
	}
	if role == model.RoleSuperAdmin || role == model.RoleAdmin {
		user.Role = role
	}
	if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}
	if restaurantID != nil {
		if *restaurantID == 0 {
			user.RestaurantID = nil
		} else {
			user.RestaurantID = restaurantID
		}
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) DeleteAdmin(id uint, requesterID uint) error {
	if id == requesterID {
		return errors.New("tidak bisa menghapus akun sendiri")
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("admin tidak ditemukan")
	}
	return s.repo.Delete(id)
}

func (s *authService) ForgotPassword(username string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", errors.New("username tidak ditemukan")
	}
	token, err := generateToken()
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	if err := s.repo.SetResetToken(user.ID, token, expiresAt); err != nil {
		return "", errors.New("gagal menyimpan token")
	}
	// Admin tidak punya field email — token dikembalikan langsung di response
	return token, nil
}

func (s *authService) Logout(id uint) error {
	return s.repo.IncrementTokenVersion(id)
}

func (s *authService) ResetPasswordWithToken(token, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	user, err := s.repo.FindByResetToken(token)
	if err != nil {
		return errors.New("token tidak valid atau sudah kadaluarsa")
	}
	if user.ResetTokenExpiresAt == nil || time.Now().UTC().After(user.ResetTokenExpiresAt.UTC()) {
		return errors.New("token sudah kadaluarsa")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.Update(&model.AdminUser{ID: user.ID, Username: user.Username, Password: string(hashed), Role: user.Role, RestaurantID: user.RestaurantID}); err != nil {
		return err
	}
	return s.repo.ClearResetToken(user.ID)
}
