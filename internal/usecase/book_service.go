package usecase

import (
	"errors"
	"strings"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/model"
)

// Sentinel errors for BookService business rule violations.
var (
	ErrBookTitleRequired       = errors.New("book title is required")
	ErrBookAuthorsRequired     = errors.New("book must have at least one author")
	ErrAuthorsCannotBeModified = errors.New("book authors cannot be modified after creation")
	ErrBookNotFound            = errors.New("book not found")
)

// BookRepository defines the persistence contract required by BookService.
// It mirrors the methods of repository.BookRepository so the concrete type
// satisfies this interface without any changes.
type BookRepository interface {
	Save(book *model.Book) (*model.Book, error)
	FindAll() ([]model.Book, error)
	FindByID(id string) (*model.Book, error)
	Update(id string, updated *model.Book) error
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
	saved, err := s.repo.Save(s.toModel(book))
	if err != nil {
		return nil, err
	}
	return s.toDomain(saved), nil
}

// GetByID retrieves a book by its unique identifier.
func (s *BookService) GetByID(id string) (*domain.Book, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toDomain(m), nil
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
			return s.toDomain(&all[i]), nil
		}
	}
	return nil, ErrBookNotFound
}

// GetAll returns every book persisted in the system.
func (s *BookService) GetAll() ([]domain.Book, error) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	books := make([]domain.Book, len(all))
	for i := range all {
		b := s.toDomain(&all[i])
		books[i] = *b
	}
	return books, nil
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

	merged := model.NewBookBuilder().
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
	return s.toDomain(&merged), nil
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
	return s.repo.Delete(book.ID)
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

// toModel converts a domain.Book to the persistence model.
func (s *BookService) toModel(b *domain.Book) *model.Book {
	m := model.NewBookBuilder().
		WithTitle(b.Title).
		WithAuthors(b.Authors).
		WithPublisher(b.Publisher).
		WithPublishedYear(b.PublishedYear).
		WithISBN10(b.ISBN10).
		WithISBN13(b.ISBN13).
		WithDescription(b.Description).
		WithGenre(b.Genre).
		Build()
	return &m
}

// toDomain converts a persistence model to a domain.Book.
func (s *BookService) toDomain(m *model.Book) *domain.Book {
	b := domain.NewBookBuilder().
		WithTitle(m.Title).
		WithAuthors(m.Authors).
		WithPublisher(m.Publisher).
		WithPublishedYear(m.PublishedYear).
		WithISBN10(m.ISBN10).
		WithISBN13(m.ISBN13).
		WithDescription(m.Description).
		WithGenre(m.Genre).
		Build()
	b.ID = m.ID.Hex()
	return &b
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
