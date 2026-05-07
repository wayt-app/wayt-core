package service

import (
	"errors"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
	"gorm.io/gorm"
)

type ReviewService interface {
	Submit(customerID, bookingID uint, rating int, comment string) (*model.Review, error)
	FindByBookingID(bookingID uint) (*model.Review, error)
	ListByRestaurant(restaurantID uint, limit, offset int) ([]model.Review, int64, error)
	StatsByRestaurant(restaurantID uint) (avgRating float64, total int64, err error)
}

type reviewService struct {
	repo        repository.ReviewRepository
	bookingRepo repository.BookingRepository
	branchRepo  repository.BranchRepository
}

func NewReviewService(repo repository.ReviewRepository, bookingRepo repository.BookingRepository, branchRepo repository.BranchRepository) ReviewService {
	return &reviewService{repo: repo, bookingRepo: bookingRepo, branchRepo: branchRepo}
}

func (s *reviewService) Submit(customerID, bookingID uint, rating int, comment string) (*model.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating harus antara 1 dan 5")
	}
	if len(comment) > 1000 {
		return nil, errors.New("komentar maksimal 1000 karakter")
	}

	booking, err := s.bookingRepo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("reservasi tidak ditemukan")
	}
	if booking.CustomerID != customerID {
		return nil, errors.New("tidak dapat memberi ulasan pada reservasi ini")
	}
	if booking.Status != model.BookingStatusCompleted {
		return nil, errors.New("ulasan hanya bisa diberikan untuk reservasi yang sudah selesai")
	}

	// Check duplicate
	if existing, err := s.repo.FindByBookingID(bookingID); err == nil && existing.ID > 0 {
		return nil, errors.New("kamu sudah memberikan ulasan untuk reservasi ini")
	}

	branch, err := s.branchRepo.FindByID(booking.BranchID)
	if err != nil {
		return nil, errors.New("data cabang tidak ditemukan")
	}

	rv := &model.Review{
		CustomerID:   customerID,
		RestaurantID: branch.RestaurantID,
		BranchID:     booking.BranchID,
		BookingID:    bookingID,
		Rating:       rating,
	}
	if comment != "" {
		rv.Comment = &comment
	}

	if err := s.repo.Create(rv); err != nil {
		return nil, errors.New("gagal menyimpan ulasan")
	}
	return rv, nil
}

func (s *reviewService) FindByBookingID(bookingID uint) (*model.Review, error) {
	rv, err := s.repo.FindByBookingID(bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rv, nil
}

func (s *reviewService) ListByRestaurant(restaurantID uint, limit, offset int) ([]model.Review, int64, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.FindByRestaurantID(restaurantID, limit, offset)
}

func (s *reviewService) StatsByRestaurant(restaurantID uint) (float64, int64, error) {
	return s.repo.StatsByRestaurantID(restaurantID)
}
