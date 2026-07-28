package usecase_test

import (
	"errors"
	"testing"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockBookRepository struct {
	mock.Mock
}

func (m *mockBookRepository) Save(book *domain.Book) (*domain.Book, error) {
	args := m.Called(book)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *mockBookRepository) FindAll() ([]domain.Book, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Book), args.Error(1)
}

func (m *mockBookRepository) FindByID(id string) (*domain.Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *mockBookRepository) Update(id string, updated *domain.Book) error {
	args := m.Called(id, updated)
	return args.Error(0)
}

func (m *mockBookRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// domainBook builds a *domain.Book with a preset ObjectID.
func domainBook(id bson.ObjectID, title string, authors []string) *domain.Book {
	b := domain.NewBookBuilder().
		WithTitle(title).
		WithAuthors(authors).
		Build()
	b.ID = id
	return &b
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestBookService_Create_Success(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	stored := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("Save", mock.AnythingOfType("*domain.Book")).Return(stored, nil)

	input := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		Build()

	got, err := svc.Create(&input)

	assert.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "Dune", got.Title)
	assert.Equal(t, []string{"Frank Herbert"}, got.Authors)
	repo.AssertExpectations(t)
}

func TestBookService_Create_WithOptionalFields(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	stored := func() *domain.Book {
		b := domain.NewBookBuilder().
			WithTitle("Dune").
			WithAuthors([]string{"Frank Herbert"}).
			WithPublisher("Chilton Books").
			WithPublishedYear(1965).
			WithGenre([]string{"Science Fiction"}).
			Build()
		b.ID = id
		return &b
	}()
	repo.On("Save", mock.AnythingOfType("*domain.Book")).Return(stored, nil)

	input := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		WithPublisher("Chilton Books").
		WithPublishedYear(1965).
		WithGenre([]string{"Science Fiction"}).
		Build()

	got, err := svc.Create(&input)

	assert.NoError(t, err)
	assert.Equal(t, "Chilton Books", got.Publisher)
	assert.Equal(t, 1965, got.PublishedYear)
	assert.Equal(t, []string{"Science Fiction"}, got.Genre)
	repo.AssertExpectations(t)
}

func TestBookService_Create_EmptyTitle(t *testing.T) {
	svc := usecase.NewBookService(&mockBookRepository{})

	input := domain.NewBookBuilder().
		WithAuthors([]string{"Frank Herbert"}).
		Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrBookTitleRequired)
}

func TestBookService_Create_BlankTitle(t *testing.T) {
	svc := usecase.NewBookService(&mockBookRepository{})

	input := domain.NewBookBuilder().
		WithTitle("   ").
		WithAuthors([]string{"Frank Herbert"}).
		Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrBookTitleRequired)
}

func TestBookService_Create_NoAuthors(t *testing.T) {
	svc := usecase.NewBookService(&mockBookRepository{})

	input := domain.NewBookBuilder().
		WithTitle("Dune").
		Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrBookAuthorsRequired)
}

func TestBookService_Create_AllBlankAuthors(t *testing.T) {
	svc := usecase.NewBookService(&mockBookRepository{})

	input := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"", "  ", "\t"}).
		Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrBookAuthorsRequired)
}

func TestBookService_Create_RepositoryError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repoErr := errors.New("connection refused")
	repo.On("Save", mock.AnythingOfType("*domain.Book")).Return(nil, repoErr)

	input := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestBookService_GetByID_Found(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	stored := domainBook(id, "1984", []string{"George Orwell"})
	repo.On("FindByID", id.Hex()).Return(stored, nil)

	got, err := svc.GetByID(id.Hex())

	assert.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "1984", got.Title)
	repo.AssertExpectations(t)
}

func TestBookService_GetByID_NotFound(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	unknownID := "000000000000000000000000"
	repo.On("FindByID", unknownID).Return(nil, errors.New("mongo: no documents in result"))

	_, err := svc.GetByID(unknownID)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestBookService_GetByID_InvalidID(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repo.On("FindByID", "bad-id").Return(nil, errors.New("invalid id"))

	_, err := svc.GetByID("bad-id")

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetByTitle
// ---------------------------------------------------------------------------

func TestBookService_GetByTitle_ExactMatch(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	all := []domain.Book{*domainBook(id, "Brave New World", []string{"Aldous Huxley"})}
	repo.On("FindAll").Return(all, nil)

	got, err := svc.GetByTitle("Brave New World")

	assert.NoError(t, err)
	assert.Equal(t, "Brave New World", got.Title)
	assert.Equal(t, id, got.ID)
	repo.AssertExpectations(t)
}

func TestBookService_GetByTitle_CaseInsensitive(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	all := []domain.Book{*domainBook(id, "Brave New World", []string{"Aldous Huxley"})}
	repo.On("FindAll").Return(all, nil)

	got, err := svc.GetByTitle("brave new world")

	assert.NoError(t, err)
	assert.Equal(t, "Brave New World", got.Title)
	repo.AssertExpectations(t)
}

func TestBookService_GetByTitle_NotFound(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repo.On("FindAll").Return([]domain.Book{}, nil)

	_, err := svc.GetByTitle("Unknown Book")

	assert.ErrorIs(t, err, usecase.ErrBookNotFound)
	repo.AssertExpectations(t)
}

func TestBookService_GetByTitle_RepositoryError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repoErr := errors.New("db down")
	repo.On("FindAll").Return(nil, repoErr)

	_, err := svc.GetByTitle("Dune")

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetAll
// ---------------------------------------------------------------------------

func TestBookService_GetAll_ReturnsList(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id1, id2 := bson.NewObjectID(), bson.NewObjectID()
	all := []domain.Book{
		*domainBook(id1, "Dune", []string{"Frank Herbert"}),
		*domainBook(id2, "1984", []string{"George Orwell"}),
	}
	repo.On("FindAll").Return(all, nil)

	got, err := svc.GetAll()

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	repo.AssertExpectations(t)
}

func TestBookService_GetAll_EmptyCollection(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repo.On("FindAll").Return([]domain.Book{}, nil)

	got, err := svc.GetAll()

	assert.NoError(t, err)
	assert.Empty(t, got)
	repo.AssertExpectations(t)
}

func TestBookService_GetAll_RepositoryError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repoErr := errors.New("timeout")
	repo.On("FindAll").Return(nil, repoErr)

	_, err := svc.GetAll()

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestBookService_Update_MutableFieldsChanged(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)
	repo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Book")).Return(nil)

	updates := domain.NewBookBuilder().
		WithTitle("Dune Messiah").
		WithPublisher("Putnam").
		WithPublishedYear(1969).
		WithGenre([]string{"Science Fiction"}).
		Build()

	got, err := svc.Update(id.Hex(), &updates)

	assert.NoError(t, err)
	assert.Equal(t, "Dune Messiah", got.Title)
	assert.Equal(t, "Putnam", got.Publisher)
	assert.Equal(t, 1969, got.PublishedYear)
	assert.Equal(t, []string{"Science Fiction"}, got.Genre)
	repo.AssertExpectations(t)
}

func TestBookService_Update_AuthorsPreservedWhenNotProvided(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)
	repo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Book")).Return(nil)

	updates := domain.NewBookBuilder().
		WithTitle("Dune Revised").
		Build()

	got, err := svc.Update(id.Hex(), &updates)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Frank Herbert"}, got.Authors)
	repo.AssertExpectations(t)
}

func TestBookService_Update_AuthorsPreservedWhenSameValueSupplied(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)
	repo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Book")).Return(nil)

	updates := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		Build()

	got, err := svc.Update(id.Hex(), &updates)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Frank Herbert"}, got.Authors)
	repo.AssertExpectations(t)
}

func TestBookService_Update_RejectsAuthorChangeByAddingCoAuthor(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)

	updates := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert", "Co-Author"}).
		Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrAuthorsCannotBeModified)
	repo.AssertNotCalled(t, "Update")
	repo.AssertExpectations(t)
}

func TestBookService_Update_RejectsAuthorChange(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)

	updates := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Someone Else"}).
		Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrAuthorsCannotBeModified)
	repo.AssertNotCalled(t, "Update")
	repo.AssertExpectations(t)
}

func TestBookService_Update_RejectsEmptyTitle(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)

	updates := domain.NewBookBuilder().
		WithTitle("   ").
		Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrBookTitleRequired)
	repo.AssertNotCalled(t, "Update")
	repo.AssertExpectations(t)
}

func TestBookService_Update_BookNotFound(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	unknownID := "000000000000000000000000"
	repo.On("FindByID", unknownID).Return(nil, errors.New("not found"))

	updates := domain.NewBookBuilder().WithTitle("Anything").Build()
	_, err := svc.Update(unknownID, &updates)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Update")
	repo.AssertExpectations(t)
}

func TestBookService_Update_RepositoryUpdateError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	existing := domainBook(id, "Dune", []string{"Frank Herbert"})
	repo.On("FindByID", id.Hex()).Return(existing, nil)

	repoErr := errors.New("write failed")
	repo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Book")).Return(repoErr)

	updates := domain.NewBookBuilder().WithTitle("New Title").Build()
	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteByID
// ---------------------------------------------------------------------------

func TestBookService_DeleteByID_Success(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	repo.On("Delete", id.Hex()).Return(nil)

	err := svc.DeleteByID(id.Hex())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBookService_DeleteByID_RepositoryError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repoErr := errors.New("delete failed")
	repo.On("Delete", "000000000000000000000000").Return(repoErr)

	err := svc.DeleteByID("000000000000000000000000")

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteByTitle
// ---------------------------------------------------------------------------

func TestBookService_DeleteByTitle_Success(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	all := []domain.Book{*domainBook(id, "The Hobbit", []string{"J.R.R. Tolkien"})}
	repo.On("FindAll").Return(all, nil)
	repo.On("Delete", id.Hex()).Return(nil)

	err := svc.DeleteByTitle("The Hobbit")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBookService_DeleteByTitle_CaseInsensitive(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	all := []domain.Book{*domainBook(id, "The Hobbit", []string{"J.R.R. Tolkien"})}
	repo.On("FindAll").Return(all, nil)
	repo.On("Delete", id.Hex()).Return(nil)

	err := svc.DeleteByTitle("the hobbit")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBookService_DeleteByTitle_NotFound(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repo.On("FindAll").Return([]domain.Book{}, nil)

	err := svc.DeleteByTitle("Nonexistent")

	assert.ErrorIs(t, err, usecase.ErrBookNotFound)
	repo.AssertNotCalled(t, "Delete")
	repo.AssertExpectations(t)
}

func TestBookService_DeleteByTitle_FindAllError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	repoErr := errors.New("db error")
	repo.On("FindAll").Return(nil, repoErr)

	err := svc.DeleteByTitle("Dune")

	assert.ErrorIs(t, err, repoErr)
	repo.AssertNotCalled(t, "Delete")
	repo.AssertExpectations(t)
}

func TestBookService_DeleteByTitle_DeleteError(t *testing.T) {
	repo := &mockBookRepository{}
	svc := usecase.NewBookService(repo)

	id := bson.NewObjectID()
	all := []domain.Book{*domainBook(id, "The Hobbit", []string{"J.R.R. Tolkien"})}
	repo.On("FindAll").Return(all, nil)

	repoErr := errors.New("delete failed")
	repo.On("Delete", id.Hex()).Return(repoErr)

	err := svc.DeleteByTitle("The Hobbit")

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}
