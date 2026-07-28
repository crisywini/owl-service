package usecase

import (
	"errors"
	"strings"

	"github.com/crisywini/owl-service/internal/domain"
)

// Sentinel errors exposed by ReviewService.
var (
	ErrReviewRateRequired        = errors.New("review rate is required")
	ErrReviewDescriptionRequired = errors.New("review description is required")
	ErrReviewBookNotFound        = errors.New("book not found for review")
	ErrReviewBookIDImmutable     = errors.New("review book reference cannot be changed")
	ErrPhrasesRequired           = errors.New("at least one phrase is required")
)

type ReviewServicePort interface {
	Create(review *domain.Review) (*domain.Review, error)
	GetByID(id string) (*domain.Review, error)
	GetByBookID(bookID string) ([]domain.Review, error)
	GetAll() ([]domain.Review, error)
	Update(id string, updates *domain.Review) (*domain.Review, error)
	AddPhrases(id string, phrases []string) (*domain.Review, error)
	DeleteByID(id string) error
}

// ReviewRepository defines the persistence contract required by ReviewService.
// The concrete repository.ReviewRepository satisfies this interface directly.
type ReviewRepository interface {
	Save(review *domain.Review) (*domain.Review, error)
	FindAll() ([]domain.Review, error)
	FindByID(id string) (*domain.Review, error)
	FindByBookID(bookID string) ([]domain.Review, error)
	Update(id string, updated *domain.Review) error
	Delete(id string) error
}

// BookFinder is the minimal interface ReviewService needs to verify a book exists.
// The concrete repository.BookRepository satisfies this interface directly.
type BookFinder interface {
	FindByID(id string) (*domain.Book, error)
}

// ReviewService enforces business rules for review operations.
type ReviewService struct {
	reviewRepo ReviewRepository
	bookFinder BookFinder
}

// NewReviewService creates a ReviewService backed by the given repositories.
func NewReviewService(reviewRepo ReviewRepository, bookFinder BookFinder) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, bookFinder: bookFinder}
}

// Create validates and persists a new review.
//
// Business rules enforced here:
//   - The referenced book must exist.
//   - Rate must be greater than zero.
//   - Description must be non-empty.
//   - StartDate, FinishDate and FavoritePhrases are optional.
func (s *ReviewService) Create(review *domain.Review) (*domain.Review, error) {
	if _, err := s.bookFinder.FindByID(review.BookID.Hex()); err != nil {
		return nil, ErrReviewBookNotFound
	}
	if err := s.validateReview(review); err != nil {
		return nil, err
	}
	return s.reviewRepo.Save(review)
}

// GetByID retrieves a review by its unique identifier.
func (s *ReviewService) GetByID(id string) (*domain.Review, error) {
	return s.reviewRepo.FindByID(id)
}

// GetByBookID returns every review associated with the given book identifier.
func (s *ReviewService) GetByBookID(bookID string) ([]domain.Review, error) {
	return s.reviewRepo.FindByBookID(bookID)
}

// GetAll returns every review persisted in the system.
func (s *ReviewService) GetAll() ([]domain.Review, error) {
	return s.reviewRepo.FindAll()
}

// Update applies mutable-field changes to an existing review.
//
// Business rules enforced here:
//   - The review must exist (repository error is propagated otherwise).
//   - Rate must remain greater than zero.
//   - Description must remain non-empty.
//   - BookID is immutable: supplying a different BookID returns ErrReviewBookIDImmutable.
func (s *ReviewService) Update(id string, updates *domain.Review) (*domain.Review, error) {
	existing, err := s.reviewRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !updates.BookID.IsZero() && updates.BookID != existing.BookID {
		return nil, ErrReviewBookIDImmutable
	}

	if err := s.validateReview(updates); err != nil {
		return nil, err
	}

	merged := domain.NewReviewBuilder(existing.BookID).
		WithRate(updates.Rate).
		WithDescription(updates.Description).
		WithStartDate(updates.StartDate).
		WithFinishDate(updates.FinishDate).
		WithFavoritePhrases(updates.FavoritePhrases).
		Build()

	if err := s.reviewRepo.Update(id, &merged); err != nil {
		return nil, err
	}

	merged.ID = existing.ID
	return &merged, nil
}

// AddPhrases appends one or more favorite phrases to an existing review.
// Returns ErrPhrasesRequired when the phrases slice is empty.
func (s *ReviewService) AddPhrases(id string, phrases []string) (*domain.Review, error) {
	if len(phrases) == 0 {
		return nil, ErrPhrasesRequired
	}

	existing, err := s.reviewRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	existing.FavoritePhrases = append(existing.FavoritePhrases, phrases...)

	if err := s.reviewRepo.Update(id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteByID removes the review with the given identifier.
func (s *ReviewService) DeleteByID(id string) error {
	return s.reviewRepo.Delete(id)
}

// validateReview enforces the mandatory-field contract for a review.
func (s *ReviewService) validateReview(review *domain.Review) error {
	if review.Rate <= 0 {
		return ErrReviewRateRequired
	}
	if strings.TrimSpace(review.Description) == "" {
		return ErrReviewDescriptionRequired
	}
	return nil
}
