package usecase

import (
	"errors"
	"strings"

	"github.com/crisywini/owl-service/internal/domain"
)

// Sentinel errors exposed by BookService.
var (
	ErrBookTitleRequired       = errors.New("book title is required")
	ErrBookAuthorsRequired     = errors.New("book must have at least one author")
	ErrAuthorsCannotBeModified = errors.New("book authors cannot be modified after creation")
	ErrBookNotFound            = errors.New("book not found")
)

// BookServicePort is the contract that consumers (e.g. HTTP handlers) depend on.
type BookServicePort interface {
	Create(book *domain.Book) (*domain.Book, error)
	GetByID(id string) (*domain.Book, error)
	GetAll() ([]domain.Book, error)
	Update(id string, updates *domain.Book) (*domain.Book, error)
	DeleteByID(id string) error
}

type BookRepository interface {
	Save(book *domain.Book) (*domain.Book, error)
	FindAll() ([]domain.Book, error)
	FindByID(id string) (*domain.Book, error)
	Update(id string, updated *domain.Book) error
	Delete(id string) error
}

// BookService enforces business rules for book operations.
type BookService struct {
	repo BookRepository
}

// NewBookService creates a BookService backed by the given repository.
func NewBookService(repo BookRepository) *BookService {
	return &BookService{repo: repo}
}

// Create validates and persists a new book.
// Both Title and at least one non-empty Author are mandatory.
// Genre, Publisher, PublishedYear, ISBN10, ISBN13 and Description are optional.
func (s *BookService) Create(book *domain.Book) (*domain.Book, error) {
	if err := s.validateBook(book); err != nil {
		return nil, err
	}
	return s.repo.Save(book)
}

// GetByID retrieves a book by its unique identifier.
func (s *BookService) GetByID(id string) (*domain.Book, error) {
	return s.repo.FindByID(id)
}

// GetByTitle returns the first book whose Title matches (case-insensitive).
// Returns ErrBookNotFound when no match exists.
func (s *BookService) GetByTitle(title string) (*domain.Book, error) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(title))
	for i := range all {
		if strings.ToLower(all[i].Title) == needle {
			return &all[i], nil
		}
	}
	return nil, ErrBookNotFound
}

// GetAll returns every book persisted in the system.
func (s *BookService) GetAll() ([]domain.Book, error) {
	return s.repo.FindAll()
}

// Update applies mutable-field changes to an existing book.
//
// Business rules enforced here:
//   - The book must exist (repository error is propagated otherwise).
//   - Title must remain non-empty.
//   - Authors are immutable: if the caller supplies authors that differ from the
//     stored ones, ErrAuthorsCannotBeModified is returned.
//   - When no authors are supplied in updates, the stored authors are kept.
//   - Genre, Publisher, PublishedYear, ISBN10, ISBN13 and Description may be
//     freely changed or cleared.
func (s *BookService) Update(id string, updates *domain.Book) (*domain.Book, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if len(updates.Authors) > 0 && !equalAuthors(existing.Authors, updates.Authors) {
		return nil, ErrAuthorsCannotBeModified
	}

	if strings.TrimSpace(updates.Title) == "" {
		return nil, ErrBookTitleRequired
	}

	merged := domain.NewBookBuilder().
		WithTitle(updates.Title).
		WithAuthors(existing.Authors).
		WithPublisher(updates.Publisher).
		WithPublishedYear(updates.PublishedYear).
		WithISBN10(updates.ISBN10).
		WithISBN13(updates.ISBN13).
		WithDescription(updates.Description).
		WithGenre(updates.Genre).
		Build()

	if err := s.repo.Update(id, &merged); err != nil {
		return nil, err
	}

	merged.ID = existing.ID
	return &merged, nil
}

// DeleteByID removes the book with the given identifier.
func (s *BookService) DeleteByID(id string) error {
	return s.repo.Delete(id)
}

// DeleteByTitle removes the first book whose Title matches (case-insensitive).
// Returns ErrBookNotFound when no match exists.
func (s *BookService) DeleteByTitle(title string) error {
	book, err := s.GetByTitle(title)
	if err != nil {
		return err
	}
	return s.repo.Delete(book.ID.Hex())
}

// validateBook enforces the mandatory-field contract for a book.
// At least one author entry must be non-blank.
func (s *BookService) validateBook(book *domain.Book) error {
	if strings.TrimSpace(book.Title) == "" {
		return ErrBookTitleRequired
	}
	for _, a := range book.Authors {
		if strings.TrimSpace(a) != "" {
			return nil
		}
	}
	return ErrBookAuthorsRequired
}

// equalAuthors reports whether two author slices have identical content in the same order.
func equalAuthors(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
