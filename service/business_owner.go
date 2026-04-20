package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wayt/wayt-core/model"
	"github.com/wayt/wayt-core/repository"
	"github.com/wayt/wayt-core/pkg/email"
	"github.com/wayt/wayt-core/pkg/whatsapp"
	"golang.org/x/crypto/bcrypt"
)

type BusinessOwnerService interface {
	Register(name, emailAddr, phone, password string) (*model.BusinessOwner, error)
	Login(emailAddr, password string) (string, error)
	FindByID(id uint) (*model.BusinessOwner, error)
	ChangePassword(id uint, currentPassword, newPassword string) error
	ForgotPassword(emailAddr string) error
	ResetPasswordWithToken(token, newPassword string) error
	VerifyEmail(token string) error
	GetSubscription(ownerID uint) (*model.Subscription, error)
	CheckBranchLimit(ownerID uint) (bool, error)
	IncrementReservation(ownerID uint) error
	List() ([]model.BusinessOwner, error)
	ListWithSubscriptions() ([]ownerWithSub, error)
	AdminApprove(subscriptionID uint, planID uint) error
	AdminReject(subscriptionID uint, notes string) error
	AdminAssignPlan(ownerID uint, planID uint) error
	Logout(id uint) error
	LoginOrRegisterWithGoogle(googleID, email, name, avatarURL string) (string, error)
	GetGoogleOAuthURL(state string) string
	ExchangeGoogleCode(code string) (googleID, email, name, avatarURL string, err error)
}

type businessOwnerService struct {
	repo                 repository.BusinessOwnerRepository
	subRepo              repository.SubscriptionRepository
	planRepo             repository.PlanRepository
	restaurantRepo       repository.RestaurantRepository
	branchRepo           repository.BranchRepository
	bookingRepo          repository.BookingRepository
	jwtSecret            []byte
	emailSvc             email.Sender
	waSender             whatsapp.Sender
	appBaseURL           string
	googleClientID       string
	googleClientSecret   string
	googleRedirectURL    string
}

func NewBusinessOwnerService(
	repo repository.BusinessOwnerRepository,
	subRepo repository.SubscriptionRepository,
	planRepo repository.PlanRepository,
	restaurantRepo repository.RestaurantRepository,
	branchRepo repository.BranchRepository,
	bookingRepo repository.BookingRepository,
	jwtSecret string,
	emailSvc email.Sender,
	waSender whatsapp.Sender,
	appBaseURL string,
	googleClientID string,
	googleClientSecret string,
	googleRedirectURL string,
) BusinessOwnerService {
	return &businessOwnerService{
		repo:               repo,
		subRepo:            subRepo,
		planRepo:           planRepo,
		restaurantRepo:     restaurantRepo,
		branchRepo:         branchRepo,
		bookingRepo:        bookingRepo,
		jwtSecret:          []byte(jwtSecret),
		emailSvc:           emailSvc,
		waSender:           waSender,
		appBaseURL:         appBaseURL,
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		googleRedirectURL:  googleRedirectURL,
	}
}

func (s *businessOwnerService) Register(name, emailAddr, phone, password string) (*model.BusinessOwner, error) {
	if name == "" || emailAddr == "" || password == "" {
		return nil, errors.New("nama, email, dan password wajib diisi")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	owner := &model.BusinessOwner{
		Name:       name,
		Email:      emailAddr,
		Phone:      phone,
		Password:   string(hashed),
		IsVerified: false,
	}
	if err := s.repo.Create(owner); err != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Start trial subscription with Starter plan
	starterPlan, err := s.planRepo.FindFirst()
	if err != nil {
		return nil, errors.New("gagal mendapatkan paket starter")
	}
	now := time.Now()
	trialEnds := now.Add(14 * 24 * time.Hour)
	sub := &model.Subscription{
		BusinessOwnerID:       owner.ID,
		PlanID:                starterPlan.ID,
		Status:                model.SubscriptionStatusTrial,
		TrialStartedAt:        &now,
		TrialEndsAt:           &trialEnds,
		ReservationsThisMonth: 0,
		LastResetAt:           now,
	}
	if err := s.subRepo.Create(sub); err != nil {
		return nil, errors.New("gagal membuat langganan trial")
	}

	// Send verification email
	verifyToken, err := generateLongToken()
	if err == nil {
		if err := s.repo.SetVerificationToken(owner.ID, verifyToken); err == nil {
			link := fmt.Sprintf("%s/api/owner/verify-email?token=%s", s.appBaseURL, verifyToken)
			html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Terima kasih telah mendaftar di Wayt. Klik link berikut untuk mengaktifkan akun bisnis Anda:</p>
<p><a href="%s" style="display:inline-block;padding:12px 24px;background:#7c3aed;color:#fff;border-radius:8px;text-decoration:none;font-weight:bold">Verifikasi Email</a></p>
<p>Atau copy link ini ke browser: <br><a href="%s">%s</a></p>
<p>Akun Anda otomatis mendapat trial 14 hari dengan paket %s.</p>
<p>Link berlaku selama 24 jam. Jika Anda tidak mendaftar, abaikan email ini.</p>
`, owner.Name, link, link, link, starterPlan.Name)
			if err := s.emailSvc.Send(emailAddr, "Verifikasi Email — Wayt Business", html); err != nil {
					log.Printf("[EMAIL ERROR] verifikasi owner %s: %v", emailAddr, err)
				}
		}
	}
	return owner, nil
}

func (s *businessOwnerService) Login(emailAddr, password string) (string, error) {
	owner, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		return "", errors.New("email atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(owner.Password), []byte(password)); err != nil {
		return "", errors.New("email atau password salah")
	}
	if !owner.IsVerified {
		return "", errors.New("email belum diverifikasi, cek inbox Anda")
	}

	// Find restaurant_id if owner has a restaurant
	var restaurantID uint
	if rest, restErr := s.restaurantRepo.FindByOwnerID(owner.ID); restErr == nil {
		restaurantID = rest.ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":           owner.ID,
		"name":          owner.Name,
		"email":         owner.Email,
		"type":          "owner",
		"restaurant_id": restaurantID,
		"token_version": owner.TokenVersion,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}

func (s *businessOwnerService) FindByID(id uint) (*model.BusinessOwner, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("business owner tidak ditemukan")
	}
	return o, nil
}

func (s *businessOwnerService) ChangePassword(id uint, currentPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	o, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("business owner tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(o.Password), []byte(currentPassword)); err != nil {
		return errors.New("password saat ini salah")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, string(hashed))
}

func (s *businessOwnerService) ForgotPassword(emailAddr string) error {
	o, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		return nil // don't leak if email exists
	}
	token, err := generateToken()
	if err != nil {
		return errors.New("gagal membuat token")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	if err := s.repo.SetResetToken(o.ID, token, expiresAt); err != nil {
		return errors.New("gagal menyimpan token")
	}
	html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Anda menerima permintaan reset password untuk akun Wayt Business Anda.</p>
<p>Token reset password Anda (berlaku 15 menit):</p>
<p style="font-size:24px;font-weight:bold;letter-spacing:4px;color:#7c3aed">%s</p>
<p>Masukkan token ini di halaman reset password. Jika Anda tidak meminta reset password, abaikan email ini.</p>
`, o.Name, token)
	if err := s.emailSvc.Send(emailAddr, "Reset Password — Wayt Business", html); err != nil {
		log.Printf("[EMAIL ERROR] reset password owner %s: %v", emailAddr, err)
	}
	return nil
}

func (s *businessOwnerService) ResetPasswordWithToken(token, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	o, err := s.repo.FindByResetToken(token)
	if err != nil {
		return errors.New("token tidak valid atau sudah kadaluarsa")
	}
	if o.ResetTokenExpiresAt == nil || time.Now().UTC().After(o.ResetTokenExpiresAt.UTC()) {
		return errors.New("token sudah kadaluarsa")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(o.ID, string(hashed)); err != nil {
		return err
	}
	return s.repo.ClearResetToken(o.ID)
}

func (s *businessOwnerService) VerifyEmail(token string) error {
	o, err := s.repo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("kode verifikasi tidak valid")
	}
	if o.IsVerified {
		return errors.New("email sudah terverifikasi")
	}
	return s.repo.MarkVerified(o.ID)
}

func (s *businessOwnerService) GetSubscription(ownerID uint) (*model.Subscription, error) {
	sub, err := s.subRepo.FindByOwnerID(ownerID)
	if err != nil {
		return nil, errors.New("langganan tidak ditemukan")
	}
	return sub, nil
}

func (s *businessOwnerService) CheckBranchLimit(ownerID uint) (bool, error) {
	sub, err := s.subRepo.FindByOwnerID(ownerID)
	if err != nil {
		return false, errors.New("langganan tidak ditemukan")
	}
	if sub.Plan == nil {
		return false, errors.New("data paket tidak lengkap")
	}
	if sub.Plan.MaxBranches == -1 {
		return true, nil // unlimited
	}

	// Find restaurant of owner
	rest, err := s.restaurantRepo.FindByOwnerID(ownerID)
	if err != nil {
		return true, nil // no restaurant yet, can create
	}

	branches, err := s.branchRepo.FindActiveByRestaurant(rest.ID)
	if err != nil {
		return false, err
	}
	return len(branches) < sub.Plan.MaxBranches, nil
}

func (s *businessOwnerService) IncrementReservation(ownerID uint) error {
	sub, err := s.subRepo.FindByOwnerID(ownerID)
	if err != nil {
		return nil // no subscription, allow
	}

	// Check if reset needed
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	if sub.LastResetAt.Before(monthStart) {
		_ = s.subRepo.ResetMonthlyCount(sub.ID)
		sub.ReservationsThisMonth = 0
		sub.WarningSent = false
	}

	if err := s.subRepo.IncrementReservations(sub.ID); err != nil {
		return err
	}

	// Check warning threshold
	if sub.Plan == nil {
		return nil
	}
	newCount := sub.ReservationsThisMonth + 1
	if sub.Plan.MaxReservationsPerMonth > 0 && !sub.WarningSent {
		threshold := int(float64(sub.Plan.MaxReservationsPerMonth) * float64(sub.Plan.WarningThresholdPct) / 100)
		if newCount >= threshold {
			owner, err := s.repo.FindByID(ownerID)
			if err == nil {
				s.sendWarningNotif(owner, sub, newCount)
			}
			// Mark warning sent
			sub.WarningSent = true
			_ = s.subRepo.Update(sub)
		}
	}
	return nil
}

func (s *businessOwnerService) sendWarningNotif(owner *model.BusinessOwner, sub *model.Subscription, current int) {
	if sub.Plan == nil {
		return
	}
	html := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Pemberitahuan: reservasi bulan ini Anda telah mencapai <strong>%d dari %d</strong> (%d%%).</p>
<p>Pertimbangkan untuk upgrade paket agar bisa menerima lebih banyak reservasi.</p>
`, owner.Name, current, sub.Plan.MaxReservationsPerMonth, sub.Plan.WarningThresholdPct)
	if err := s.emailSvc.Send(owner.Email, "Peringatan Kuota Reservasi — Wayt Business", html); err != nil {
		log.Printf("[EMAIL ERROR] peringatan kuota owner %s: %v", owner.Email, err)
	}

	if sub.Plan.WaNotifEnabled && owner.Phone != "" {
		msg := fmt.Sprintf(
			"Halo *%s*! Reservasi bulan ini sudah %d dari %d (%d%%). Pertimbangkan upgrade paket.",
			owner.Name, current, sub.Plan.MaxReservationsPerMonth, sub.Plan.WarningThresholdPct,
		)
		if err := s.waSender.Send(owner.Phone, msg); err != nil {
			log.Printf("[WA ERROR] peringatan kuota owner %s: %v", owner.Phone, err)
		}
	}
}

type ownerWithSub struct {
	ID           uint                `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Phone        string              `json:"phone"`
	IsVerified   bool                `json:"is_verified"`
	Subscription *model.Subscription `json:"subscription"`
}

func (s *businessOwnerService) List() ([]model.BusinessOwner, error) {
	return s.repo.List()
}

func (s *businessOwnerService) ListWithSubscriptions() ([]ownerWithSub, error) {
	owners, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	var result []ownerWithSub
	for _, o := range owners {
		item := ownerWithSub{ID: o.ID, Name: o.Name, Email: o.Email, Phone: o.Phone, IsVerified: o.IsVerified}
		sub, err := s.subRepo.FindByOwnerID(o.ID)
		if err == nil {
			item.Subscription = sub
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *businessOwnerService) AdminApprove(subscriptionID uint, planID uint) error {
	sub, err := s.subRepo.FindByID(subscriptionID)
	if err != nil {
		return errors.New("langganan tidak ditemukan")
	}
	if planID > 0 {
		if _, err := s.planRepo.FindByID(planID); err != nil {
			return errors.New("paket tidak ditemukan")
		}
		sub.PlanID = planID
	}
	sub.Status = model.SubscriptionStatusActive
	now := time.Now()
	sub.ActivatedAt = &now
	return s.subRepo.Update(sub)
}

func (s *businessOwnerService) AdminReject(subscriptionID uint, notes string) error {
	_, err := s.subRepo.FindByID(subscriptionID)
	if err != nil {
		return errors.New("langganan tidak ditemukan")
	}
	return s.subRepo.UpdateStatus(subscriptionID, model.SubscriptionStatusSuspended, notes)
}

func (s *businessOwnerService) AdminAssignPlan(ownerID uint, planID uint) error {
	sub, err := s.subRepo.FindByOwnerID(ownerID)
	if err != nil {
		return errors.New("langganan tidak ditemukan")
	}
	plan, err := s.planRepo.FindByID(planID)
	if err != nil {
		return errors.New("paket tidak ditemukan")
	}
	now := time.Now()
	sub.PlanID = planID
	sub.Status = model.SubscriptionStatusActive
	sub.ActivatedAt = &now
	if err := s.subRepo.Update(sub); err != nil {
		return err
	}
	applyOverLimitFlags(s.bookingRepo, ownerID, plan.MaxReservationsPerMonth, sub.ReservationsThisMonth)
	return nil
}

func (s *businessOwnerService) Logout(id uint) error {
	return s.repo.IncrementTokenVersion(id)
}

func (s *businessOwnerService) GetGoogleOAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.googleClientID)
	params.Set("redirect_uri", s.googleRedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	params.Set("access_type", "online")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (s *businessOwnerService) ExchangeGoogleCode(code string) (googleID, emailAddr, name, avatarURL string, err error) {
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

func (s *businessOwnerService) LoginOrRegisterWithGoogle(googleID, emailAddr, name, avatarURL string) (string, error) {
	owner, err := s.repo.FindByGoogleID(googleID)
	if err != nil {
		// Not found by google_id — try by email
		owner, err = s.repo.FindByEmail(emailAddr)
		if err != nil {
			// New user: register them
			randomPw := make([]byte, 32)
			if _, err2 := rand.Read(randomPw); err2 != nil {
				return "", errors.New("gagal membuat akun")
			}
			hashed, _ := bcrypt.GenerateFromPassword(randomPw, bcrypt.DefaultCost)
			owner = &model.BusinessOwner{
				Name:       name,
				Email:      emailAddr,
				Password:   string(hashed),
				IsVerified: true,
			}
			if err3 := s.repo.Create(owner); err3 != nil {
				return "", errors.New("gagal membuat akun")
			}
			// Start trial subscription
			starterPlan, err4 := s.planRepo.FindFirst()
			if err4 == nil {
				now := time.Now()
				trialEnds := now.Add(14 * 24 * time.Hour)
				sub := &model.Subscription{
					BusinessOwnerID: owner.ID,
					PlanID:          starterPlan.ID,
					Status:          model.SubscriptionStatusTrial,
					TrialStartedAt:  &now,
					TrialEndsAt:     &trialEnds,
					LastResetAt:     now,
				}
				_ = s.subRepo.Create(sub)
			}
		}
		// Link google_id to existing account
		_ = s.repo.SetGoogleInfo(owner.ID, googleID, avatarURL)
	}

	var restaurantID uint
	if rest, restErr := s.restaurantRepo.FindByOwnerID(owner.ID); restErr == nil {
		restaurantID = rest.ID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":           owner.ID,
		"name":          owner.Name,
		"email":         owner.Email,
		"type":          "owner",
		"restaurant_id": restaurantID,
		"token_version": owner.TokenVersion,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}
	return signed, nil
}
