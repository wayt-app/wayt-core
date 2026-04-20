package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wayt/wayt-core/internal/model"
	"github.com/wayt/wayt-core/internal/repository"
	"github.com/wayt/wayt-core/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

type CustomerService interface {
	Register(name, email, phone, password string) (*model.Customer, error)
	Login(email, password string) (string, error)
	FindByID(id uint) (*model.Customer, error)
	ChangePassword(id uint, currentPassword, newPassword string) error
	ResetPassword(id uint, newPassword string) error
	List() ([]model.Customer, error)
	ForgotPassword(email string) error
	ResetPasswordWithToken(token, newPassword string) error
	VerifyEmail(token string) error
	Logout(id uint) error
	UpdateProfile(id uint, name, phone string) error
	GetGoogleOAuthURL(state string) string
	ExchangeGoogleCode(code string) (googleID, email, name, avatarURL string, err error)
	LoginOrRegisterWithGoogle(googleID, email, name, avatarURL string) (string, error)
}

type customerService struct {
	repo                repository.CustomerRepository
	jwtSecret           []byte
	emailSvc            email.Sender
	appBaseURL          string
	googleClientID      string
	googleClientSecret  string
	googleRedirectURL   string
}

func NewCustomerService(repo repository.CustomerRepository, jwtSecret string, emailSvc email.Sender, appBaseURL, googleClientID, googleClientSecret, googleRedirectURL string) CustomerService {
	return &customerService{
		repo:               repo,
		jwtSecret:          []byte(jwtSecret),
		emailSvc:           emailSvc,
		appBaseURL:         appBaseURL,
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		googleRedirectURL:  googleRedirectURL,
	}
}

func (s *customerService) Register(name, emailAddr, phone, password string) (*model.Customer, error) {
	if name == "" || emailAddr == "" || phone == "" || password == "" {
		return nil, errors.New("semua field wajib diisi")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	c := &model.Customer{Name: name, Email: emailAddr, Phone: phone, Password: string(hashed), IsVerified: false}
	if err := s.repo.Create(c); err != nil {
		return nil, errors.New("email sudah terdaftar")
	}
	// Send verification email with link
	verifyToken, err := generateLongToken()
	if err == nil {
		if err := s.repo.SetVerificationToken(c.ID, verifyToken); err == nil {
			link := fmt.Sprintf("%s/api/customers/verify-email?token=%s", s.appBaseURL, verifyToken)
			html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Terima kasih telah mendaftar di Wayt. Klik link berikut untuk mengaktifkan akun Anda:</p>
<p><a href="%s" style="display:inline-block;padding:12px 24px;background:#16a34a;color:#fff;border-radius:8px;text-decoration:none;font-weight:bold">Verifikasi Email</a></p>
<p>Atau copy link ini ke browser: <br><a href="%s">%s</a></p>
<p>Link berlaku selama 24 jam. Jika Anda tidak mendaftar, abaikan email ini.</p>
`, c.Name, link, link, link)
			if err := s.emailSvc.Send(emailAddr, "Verifikasi Email — Wayt", html); err != nil {
					log.Printf("[EMAIL ERROR] verifikasi customer %s: %v", emailAddr, err)
				}
		}
	}
	return c, nil
}

func (s *customerService) Login(email, password string) (string, error) {
	c, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", errors.New("email atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password)); err != nil {
		return "", errors.New("email atau password salah")
	}
	if !c.IsVerified {
		return "", errors.New("email belum diverifikasi, cek inbox Anda")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":           c.ID,
		"name":          c.Name,
		"email":         c.Email,
		"type":          "customer",
		"token_version": c.TokenVersion,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}

func (s *customerService) FindByID(id uint) (*model.Customer, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("customer tidak ditemukan")
	}
	return c, nil
}

func (s *customerService) ChangePassword(id uint, currentPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	c, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("customer tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(currentPassword)); err != nil {
		return errors.New("password saat ini salah")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, string(hashed))
}

func (s *customerService) ResetPassword(id uint, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, string(hashed))
}

func (s *customerService) List() ([]model.Customer, error) {
	return s.repo.List()
}

func (s *customerService) VerifyEmail(token string) error {
	c, err := s.repo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("kode verifikasi tidak valid")
	}
	if c.IsVerified {
		return errors.New("email sudah terverifikasi")
	}
	return s.repo.MarkVerified(c.ID)
}

func (s *customerService) ForgotPassword(emailAddr string) error {
	c, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		// Jangan bocorkan apakah email terdaftar
		return nil
	}
	token, err := generateToken()
	if err != nil {
		return errors.New("gagal membuat token")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	if err := s.repo.SetResetToken(c.ID, token, expiresAt); err != nil {
		return errors.New("gagal menyimpan token")
	}
	html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Anda menerima permintaan reset password untuk akun Wayt Anda.</p>
<p>Token reset password Anda (berlaku 15 menit):</p>
<p style="font-size:24px;font-weight:bold;letter-spacing:4px;color:#16a34a">%s</p>
<p>Masukkan token ini di halaman reset password. Jika Anda tidak meminta reset password, abaikan email ini.</p>
`, c.Name, token)
	if err := s.emailSvc.Send(emailAddr, "Reset Password — Wayt", html); err != nil {
		log.Printf("[EMAIL ERROR] reset password customer %s: %v", emailAddr, err)
	}
	return nil
}

func (s *customerService) ResetPasswordWithToken(token, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	c, err := s.repo.FindByResetToken(token)
	if err != nil {
		return errors.New("token tidak valid atau sudah kadaluarsa")
	}
	if c.ResetTokenExpiresAt == nil || time.Now().UTC().After(c.ResetTokenExpiresAt.UTC()) {
		return errors.New("token sudah kadaluarsa")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(c.ID, string(hashed)); err != nil {
		return err
	}
	return s.repo.ClearResetToken(c.ID)
}

func (s *customerService) Logout(id uint) error {
	return s.repo.IncrementTokenVersion(id)
}

func (s *customerService) UpdateProfile(id uint, name, phone string) error {
	if name == "" {
		return errors.New("nama tidak boleh kosong")
	}
	if len(name) > 100 {
		return errors.New("nama maksimal 100 karakter")
	}
	if len(phone) > 20 {
		return errors.New("nomor telepon maksimal 20 karakter")
	}
	return s.repo.UpdateProfile(id, name, phone)
}

func (s *customerService) GetGoogleOAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.googleClientID)
	params.Set("redirect_uri", s.googleRedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (s *customerService) ExchangeGoogleCode(code string) (googleID, emailAddr, name, avatarURL string, err error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {s.googleClientID},
		"client_secret": {s.googleClientSecret},
		"redirect_uri":  {s.googleRedirectURL},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return "", "", "", "", errors.New("gagal menghubungi Google")
	}
	defer resp.Body.Close()
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
		return "", "", "", "", errors.New("gagal mendapatkan token Google")
	}

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	client := &http.Client{}
	infoResp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", errors.New("gagal mendapatkan profil Google")
	}
	defer infoResp.Body.Close()
	var userInfo struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(infoResp.Body).Decode(&userInfo); err != nil || userInfo.Sub == "" {
		return "", "", "", "", errors.New("gagal membaca profil Google")
	}
	return userInfo.Sub, userInfo.Email, userInfo.Name, userInfo.Picture, nil
}

func (s *customerService) LoginOrRegisterWithGoogle(googleID, emailAddr, name, avatarURL string) (string, error) {
	c, err := s.repo.FindByGoogleID(googleID)
	if err != nil {
		// Not found by google_id — try by email
		c, err = s.repo.FindByEmail(emailAddr)
		if err != nil {
			// New user: register them
			randomPw := make([]byte, 32)
			if _, err2 := rand.Read(randomPw); err2 != nil {
				return "", errors.New("gagal membuat akun")
			}
			hashed, _ := bcrypt.GenerateFromPassword(randomPw, bcrypt.DefaultCost)
			c = &model.Customer{
				Name:       name,
				Email:      emailAddr,
				Phone:      "",
				Password:   string(hashed),
				IsVerified: true,
			}
			if err3 := s.repo.Create(c); err3 != nil {
				return "", errors.New("gagal membuat akun")
			}
		}
		// Link google_id to existing account
		_ = s.repo.SetGoogleInfo(c.ID, googleID, avatarURL)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":           c.ID,
		"name":          c.Name,
		"email":         c.Email,
		"type":          "customer",
		"token_version": c.TokenVersion,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}

// generateToken returns an 8-character uppercase alphanumeric token (easy to type).
func generateToken() (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no ambiguous chars (0/O, 1/I)
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b), nil
}

// generateLongToken returns a 32-char hex token for use in URLs (not user-typed).
func generateLongToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
