package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockReviewRepository struct {
	mock.Mock
}

func (m *mockReviewRepository) Save(review *domain.Review) (*domain.Review, error) {
	args := m.Called(review)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewRepository) FindAll() ([]domain.Review, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Review), args.Error(1)
}

func (m *mockReviewRepository) FindByID(id string) (*domain.Review, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewRepository) FindByBookID(bookID string) ([]domain.Review, error) {
	args := m.Called(bookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Review), args.Error(1)
}

func (m *mockReviewRepository) Update(id string, updated *domain.Review) error {
	args := m.Called(id, updated)
	return args.Error(0)
}

func (m *mockReviewRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockBookFinder struct {
	mock.Mock
}

func (m *mockBookFinder) FindByID(id string) (*domain.Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// domainReview builds a *domain.Review with a preset ObjectID.
func domainReview(id, bookID bson.ObjectID, rate int, description string) *domain.Review {
	r := domain.NewReviewBuilder(bookID).
		WithRate(rate).
		WithDescription(description).
		Build()
	r.ID = id
	return &r
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestReviewService_Create_Success(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	reviewID := bson.NewObjectID()

	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).
		WithRate(5).
		WithDescription("An absolute masterpiece.").
		Build()
	stored := domainReview(reviewID, bookID, 5, "An absolute masterpiece.")
	reviewRepo.On("Save", mock.AnythingOfType("*domain.Review")).Return(stored, nil)

	got, err := svc.Create(&input)

	assert.NoError(t, err)
	assert.Equal(t, reviewID, got.ID)
	assert.Equal(t, bookID, got.BookID)
	assert.Equal(t, 5, got.Rate)
	bookFinder.AssertExpectations(t)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Create_WithOptionalFields(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	reviewID := bson.NewObjectID()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	finish := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).
		WithRate(4).
		WithDescription("Thoroughly enjoyed it.").
		WithStartDate(start).
		WithFinishDate(finish).
		AddFavoritePhrase("The spice must flow.").
		Build()

	stored := func() *domain.Review {
		r := input
		r.ID = reviewID
		return &r
	}()
	reviewRepo.On("Save", mock.AnythingOfType("*domain.Review")).Return(stored, nil)

	got, err := svc.Create(&input)

	assert.NoError(t, err)
	assert.Equal(t, start, got.StartDate)
	assert.Equal(t, finish, got.FinishDate)
	assert.Equal(t, []string{"The spice must flow."}, got.FavoritePhrases)
	bookFinder.AssertExpectations(t)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Create_BookNotFound(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	bookFinder.On("FindByID", bookID.Hex()).Return(nil, errors.New("not found"))

	input := domain.NewReviewBuilder(bookID).WithRate(5).WithDescription("Great.").Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrReviewBookNotFound)
	reviewRepo.AssertNotCalled(t, "Save")
	bookFinder.AssertExpectations(t)
}

func TestReviewService_Create_ZeroRate(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).WithRate(0).WithDescription("Great.").Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrReviewRateRequired)
	reviewRepo.AssertNotCalled(t, "Save")
	bookFinder.AssertExpectations(t)
}

func TestReviewService_Create_NegativeRate(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).WithRate(-1).WithDescription("Great.").Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrReviewRateRequired)
	reviewRepo.AssertNotCalled(t, "Save")
	bookFinder.AssertExpectations(t)
}

func TestReviewService_Create_EmptyDescription(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).WithRate(4).Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrReviewDescriptionRequired)
	reviewRepo.AssertNotCalled(t, "Save")
	bookFinder.AssertExpectations(t)
}

func TestReviewService_Create_BlankDescription(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	input := domain.NewReviewBuilder(bookID).WithRate(4).WithDescription("   ").Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, usecase.ErrReviewDescriptionRequired)
	reviewRepo.AssertNotCalled(t, "Save")
	bookFinder.AssertExpectations(t)
}

func TestReviewService_Create_RepositoryError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	bookFinder := &mockBookFinder{}
	svc := usecase.NewReviewService(reviewRepo, bookFinder)

	bookID := bson.NewObjectID()
	storedBook := domainBook(bookID, "Dune", []string{"Frank Herbert"})
	bookFinder.On("FindByID", bookID.Hex()).Return(storedBook, nil)

	repoErr := errors.New("connection refused")
	reviewRepo.On("Save", mock.AnythingOfType("*domain.Review")).Return(nil, repoErr)

	input := domain.NewReviewBuilder(bookID).WithRate(3).WithDescription("It was okay.").Build()

	_, err := svc.Create(&input)

	assert.ErrorIs(t, err, repoErr)
	bookFinder.AssertExpectations(t)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestReviewService_GetByID_Found(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	stored := domainReview(id, bookID, 4, "Loved it.")
	reviewRepo.On("FindByID", id.Hex()).Return(stored, nil)

	got, err := svc.GetByID(id.Hex())

	assert.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, 4, got.Rate)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_GetByID_NotFound(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	unknownID := "000000000000000000000000"
	reviewRepo.On("FindByID", unknownID).Return(nil, errors.New("not found"))

	_, err := svc.GetByID(unknownID)

	assert.Error(t, err)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetByBookID
// ---------------------------------------------------------------------------

func TestReviewService_GetByBookID_ReturnsList(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	bookID := bson.NewObjectID()
	reviews := []domain.Review{
		*domainReview(bson.NewObjectID(), bookID, 5, "Fantastic."),
		*domainReview(bson.NewObjectID(), bookID, 4, "Really good."),
	}
	reviewRepo.On("FindByBookID", bookID.Hex()).Return(reviews, nil)

	got, err := svc.GetByBookID(bookID.Hex())

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_GetByBookID_Empty(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	bookID := bson.NewObjectID()
	reviewRepo.On("FindByBookID", bookID.Hex()).Return([]domain.Review{}, nil)

	got, err := svc.GetByBookID(bookID.Hex())

	assert.NoError(t, err)
	assert.Empty(t, got)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_GetByBookID_RepositoryError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	repoErr := errors.New("db down")
	reviewRepo.On("FindByBookID", mock.AnythingOfType("string")).Return(nil, repoErr)

	_, err := svc.GetByBookID(bson.NewObjectID().Hex())

	assert.ErrorIs(t, err, repoErr)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetAll
// ---------------------------------------------------------------------------

func TestReviewService_GetAll_ReturnsList(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	bookID := bson.NewObjectID()
	all := []domain.Review{
		*domainReview(bson.NewObjectID(), bookID, 5, "Amazing."),
		*domainReview(bson.NewObjectID(), bookID, 3, "Average."),
	}
	reviewRepo.On("FindAll").Return(all, nil)

	got, err := svc.GetAll()

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_GetAll_Empty(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	reviewRepo.On("FindAll").Return([]domain.Review{}, nil)

	got, err := svc.GetAll()

	assert.NoError(t, err)
	assert.Empty(t, got)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_GetAll_RepositoryError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	repoErr := errors.New("timeout")
	reviewRepo.On("FindAll").Return(nil, repoErr)

	_, err := svc.GetAll()

	assert.ErrorIs(t, err, repoErr)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestReviewService_Update_MutableFieldsChanged(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(nil)

	updates := domain.NewReviewBuilder(bson.ObjectID{}).
		WithRate(5).
		WithDescription("Changed my mind, it was amazing!").
		AddFavoritePhrase("A memorable quote.").
		Build()

	got, err := svc.Update(id.Hex(), &updates)

	assert.NoError(t, err)
	assert.Equal(t, 5, got.Rate)
	assert.Equal(t, "Changed my mind, it was amazing!", got.Description)
	assert.Equal(t, []string{"A memorable quote."}, got.FavoritePhrases)
	assert.Equal(t, bookID, got.BookID)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_BookIDPreservedWhenNotSupplied(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(nil)

	// Zero BookID in updates means "not supplied" — existing BookID is preserved
	updates := domain.NewReviewBuilder(bson.ObjectID{}).WithRate(4).WithDescription("Better than I thought.").Build()

	got, err := svc.Update(id.Hex(), &updates)

	assert.NoError(t, err)
	assert.Equal(t, bookID, got.BookID)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_RejectsBookIDChange(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)

	otherBookID := bson.NewObjectID()
	updates := domain.NewReviewBuilder(otherBookID).WithRate(5).WithDescription("Great!").Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrReviewBookIDImmutable)
	reviewRepo.AssertNotCalled(t, "Update")
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_RejectsZeroRate(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)

	updates := domain.NewReviewBuilder(bson.ObjectID{}).WithRate(0).WithDescription("Fine.").Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrReviewRateRequired)
	reviewRepo.AssertNotCalled(t, "Update")
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_RejectsEmptyDescription(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)

	updates := domain.NewReviewBuilder(bson.ObjectID{}).WithRate(4).Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, usecase.ErrReviewDescriptionRequired)
	reviewRepo.AssertNotCalled(t, "Update")
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_ReviewNotFound(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	unknownID := "000000000000000000000000"
	reviewRepo.On("FindByID", unknownID).Return(nil, errors.New("not found"))

	updates := domain.NewReviewBuilder(bson.ObjectID{}).WithRate(3).WithDescription("ok").Build()
	_, err := svc.Update(unknownID, &updates)

	assert.Error(t, err)
	reviewRepo.AssertNotCalled(t, "Update")
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_Update_RepositoryUpdateError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 3, "It was okay.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)

	repoErr := errors.New("write failed")
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(repoErr)

	updates := domain.NewReviewBuilder(bson.ObjectID{}).WithRate(5).WithDescription("Great!").Build()

	_, err := svc.Update(id.Hex(), &updates)

	assert.ErrorIs(t, err, repoErr)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// AddPhrases
// ---------------------------------------------------------------------------

func TestReviewService_AddPhrases_Success(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 4, "Loved it.")
	existing.FavoritePhrases = []string{"First phrase."}
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(nil)

	got, err := svc.AddPhrases(id.Hex(), []string{"Second phrase.", "Third phrase."})

	assert.NoError(t, err)
	assert.Equal(t, []string{"First phrase.", "Second phrase.", "Third phrase."}, got.FavoritePhrases)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_AddPhrases_ToEmptyList(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 4, "Loved it.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(nil)

	got, err := svc.AddPhrases(id.Hex(), []string{"First ever phrase."})

	assert.NoError(t, err)
	assert.Equal(t, []string{"First ever phrase."}, got.FavoritePhrases)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_AddPhrases_EmptySliceReturnsError(t *testing.T) {
	svc := usecase.NewReviewService(&mockReviewRepository{}, &mockBookFinder{})

	_, err := svc.AddPhrases(bson.NewObjectID().Hex(), []string{})

	assert.ErrorIs(t, err, usecase.ErrPhrasesRequired)
}

func TestReviewService_AddPhrases_ReviewNotFound(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	unknownID := "000000000000000000000000"
	reviewRepo.On("FindByID", unknownID).Return(nil, errors.New("not found"))

	_, err := svc.AddPhrases(unknownID, []string{"A phrase."})

	assert.Error(t, err)
	reviewRepo.AssertNotCalled(t, "Update")
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_AddPhrases_UpdateError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	bookID := bson.NewObjectID()
	existing := domainReview(id, bookID, 4, "Loved it.")
	reviewRepo.On("FindByID", id.Hex()).Return(existing, nil)

	repoErr := errors.New("write failed")
	reviewRepo.On("Update", id.Hex(), mock.AnythingOfType("*domain.Review")).Return(repoErr)

	_, err := svc.AddPhrases(id.Hex(), []string{"A phrase."})

	assert.ErrorIs(t, err, repoErr)
	reviewRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteByID
// ---------------------------------------------------------------------------

func TestReviewService_DeleteByID_Success(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	id := bson.NewObjectID()
	reviewRepo.On("Delete", id.Hex()).Return(nil)

	err := svc.DeleteByID(id.Hex())

	assert.NoError(t, err)
	reviewRepo.AssertExpectations(t)
}

func TestReviewService_DeleteByID_RepositoryError(t *testing.T) {
	reviewRepo := &mockReviewRepository{}
	svc := usecase.NewReviewService(reviewRepo, &mockBookFinder{})

	repoErr := errors.New("delete failed")
	reviewRepo.On("Delete", "000000000000000000000000").Return(repoErr)

	err := svc.DeleteByID("000000000000000000000000")

	assert.ErrorIs(t, err, repoErr)
	reviewRepo.AssertExpectations(t)
}
