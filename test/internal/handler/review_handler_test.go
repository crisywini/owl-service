package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/handler"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

// Compile-time assertion: *mockReviewService must implement usecase.ReviewServicePort.
var _ usecase.ReviewServicePort = (*mockReviewService)(nil)

type mockReviewService struct {
	mock.Mock
}

func (m *mockReviewService) Create(review *domain.Review) (*domain.Review, error) {
	args := m.Called(review)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewService) GetByID(id string) (*domain.Review, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewService) GetByBookID(bookID string) ([]domain.Review, error) {
	args := m.Called(bookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Review), args.Error(1)
}

func (m *mockReviewService) GetAll() ([]domain.Review, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Review), args.Error(1)
}

func (m *mockReviewService) Update(id string, updates *domain.Review) (*domain.Review, error) {
	args := m.Called(id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewService) AddPhrases(id string, phrases []string) (*domain.Review, error) {
	args := m.Called(id, phrases)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Review), args.Error(1)
}

func (m *mockReviewService) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupReviewRouter(h *handler.ReviewHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/reviews", h.PostReview)
	r.GET("/reviews", h.GetReviews)
	r.GET("/reviews/:id", h.GetReview)
	r.GET("/reviews/book/:bookId", h.GetReviewsByBook)
	r.PUT("/reviews/:id", h.PutReview)
	r.PATCH("/reviews/:id/phrases", h.PatchReviewPhrases)
	r.DELETE("/reviews/:id", h.DeleteReview)
	return r
}

func sampleReview() *domain.Review {
	bookID := bson.NewObjectID()
	review := domain.NewReviewBuilder(bookID).
		WithRate(5).
		WithDescription("An excellent read.").
		WithStartDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)).
		WithFinishDate(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)).
		WithFavoritePhrases([]string{"To be or not to be"}).
		Build()
	return &review
}

func jsonReviewBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewBuffer(b)
}

// ---------------------------------------------------------------------------
// PostReview
// ---------------------------------------------------------------------------

func TestPostReview_Success(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	saved := *review
	saved.ID = bson.NewObjectID()

	svc.On("Create", mock.AnythingOfType("*domain.Review")).Return(&saved, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, saved.ID.Hex(), resp["id"])

	svc.AssertExpectations(t)
}

func TestPostReview_MalformedJSON(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Malformed Review", resp["message"])

	svc.AssertNotCalled(t, "Create")
}

func TestPostReview_BookNotFound(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	svc.On("Create", mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewBookNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewBookNotFound.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPostReview_RateRequired(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.Rate = 0

	svc.On("Create", mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewRateRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewRateRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPostReview_DescriptionRequired(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.Description = ""

	svc.On("Create", mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewDescriptionRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewDescriptionRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPostReview_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	svc.On("Create", mock.AnythingOfType("*domain.Review")).Return(nil, errors.New("db write failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.AssertExpectations(t)
}

// TestPostReview_ContextBodyLoaded verifies the gin context deserialises the
// request body and passes a fully populated Review to the service.
func TestPostReview_ContextBodyLoaded(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	saved := *review
	saved.ID = bson.NewObjectID()

	var captured *domain.Review
	svc.On("Create", mock.AnythingOfType("*domain.Review")).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(*domain.Review)
		}).
		Return(&saved, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reviews", jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotNil(t, captured)
	assert.Equal(t, review.Rate, captured.Rate)
	assert.Equal(t, review.Description, captured.Description)
	assert.Equal(t, review.BookID, captured.BookID)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetReview
// ---------------------------------------------------------------------------

func TestGetReview_Success(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()

	svc.On("GetByID", review.ID.Hex()).Return(review, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/"+review.ID.Hex(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, review.Rate, resp.Rate)
	assert.Equal(t, review.Description, resp.Description)

	svc.AssertExpectations(t)
}

func TestGetReview_NotFound(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("GetByID", id).Return(nil, errors.New("mongo: no documents in result"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	svc.AssertExpectations(t)
}

func TestGetReview_IDPassedToService(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()
	expectedID := review.ID.Hex()

	var capturedID string
	svc.On("GetByID", expectedID).
		Run(func(args mock.Arguments) { capturedID = args.String(0) }).
		Return(review, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/"+expectedID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedID, capturedID)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetReviews
// ---------------------------------------------------------------------------

func TestGetReviews_ReturnsList(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	r1 := sampleReview()
	r1.ID = bson.NewObjectID()
	r2 := sampleReview()
	r2.ID = bson.NewObjectID()

	svc.On("GetAll").Return([]domain.Review{*r1, *r2}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)

	svc.AssertExpectations(t)
}

func TestGetReviews_EmptyList(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	svc.On("GetAll").Return([]domain.Review{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)

	svc.AssertExpectations(t)
}

func TestGetReviews_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	svc.On("GetAll").Return(nil, errors.New("db timeout"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "db timeout", resp["message"])

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetReviewsByBook
// ---------------------------------------------------------------------------

func TestGetReviewsByBook_ReturnsList(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	bookID := bson.NewObjectID().Hex()
	r1 := sampleReview()
	r1.ID = bson.NewObjectID()

	svc.On("GetByBookID", bookID).Return([]domain.Review{*r1}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/book/"+bookID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)

	svc.AssertExpectations(t)
}

func TestGetReviewsByBook_EmptyList(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	bookID := bson.NewObjectID().Hex()
	svc.On("GetByBookID", bookID).Return([]domain.Review{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/book/"+bookID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)

	svc.AssertExpectations(t)
}

func TestGetReviewsByBook_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	bookID := bson.NewObjectID().Hex()
	svc.On("GetByBookID", bookID).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/book/"+bookID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.AssertExpectations(t)
}

func TestGetReviewsByBook_BookIDPassedToService(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	bookID := bson.NewObjectID().Hex()

	var capturedBookID string
	svc.On("GetByBookID", bookID).
		Run(func(args mock.Arguments) { capturedBookID = args.String(0) }).
		Return([]domain.Review{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/book/"+bookID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, bookID, capturedBookID)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// PutReview
// ---------------------------------------------------------------------------

func TestPutReview_Success(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()
	updated := *review
	updated.Description = "Updated description."

	svc.On("Update", review.ID.Hex(), mock.AnythingOfType("*domain.Review")).Return(&updated, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+review.ID.Hex(), jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, updated.Description, resp.Description)

	svc.AssertExpectations(t)
}

func TestPutReview_MalformedJSON(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+id, bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Malformed Review", resp["message"])

	svc.AssertNotCalled(t, "Update")
}

func TestPutReview_BookIDImmutable(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	review := sampleReview()

	svc.On("Update", id, mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewBookIDImmutable)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+id, jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewBookIDImmutable.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutReview_RateRequired(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	review := sampleReview()
	review.Rate = 0

	svc.On("Update", id, mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewRateRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+id, jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewRateRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutReview_DescriptionRequired(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	review := sampleReview()
	review.Description = ""

	svc.On("Update", id, mock.AnythingOfType("*domain.Review")).Return(nil, usecase.ErrReviewDescriptionRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+id, jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrReviewDescriptionRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutReview_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	review := sampleReview()

	svc.On("Update", id, mock.AnythingOfType("*domain.Review")).Return(nil, errors.New("db write failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+id, jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.AssertExpectations(t)
}

// TestPutReview_ContextBodyLoaded verifies the request body is fully deserialised
// and the correct ID is extracted from the URL parameter.
func TestPutReview_ContextBodyLoaded(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()
	expectedID := review.ID.Hex()

	var capturedID string
	var capturedReview *domain.Review
	svc.On("Update", expectedID, mock.AnythingOfType("*domain.Review")).
		Run(func(args mock.Arguments) {
			capturedID = args.String(0)
			capturedReview = args.Get(1).(*domain.Review)
		}).
		Return(review, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+expectedID, jsonReviewBody(t, review))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedID, capturedID)
	assert.NotNil(t, capturedReview)
	assert.Equal(t, review.Rate, capturedReview.Rate)
	assert.Equal(t, review.Description, capturedReview.Description)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// PatchReviewPhrases
// ---------------------------------------------------------------------------

func TestPatchReviewPhrases_Success(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()
	phrases := []string{"new phrase one", "new phrase two"}

	updated := *review
	updated.FavoritePhrases = append(updated.FavoritePhrases, phrases...)

	svc.On("AddPhrases", review.ID.Hex(), phrases).Return(&updated, nil)

	body, _ := json.Marshal(map[string][]string{"phrases": phrases})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/reviews/"+review.ID.Hex()+"/phrases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.Review
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp.FavoritePhrases, "new phrase one")

	svc.AssertExpectations(t)
}

func TestPatchReviewPhrases_MalformedJSON(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/reviews/"+id+"/phrases", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Malformed request", resp["message"])

	svc.AssertNotCalled(t, "AddPhrases")
}

func TestPatchReviewPhrases_PhrasesRequired(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("AddPhrases", id, []string{}).Return(nil, usecase.ErrPhrasesRequired)

	body, _ := json.Marshal(map[string][]string{"phrases": {}})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/reviews/"+id+"/phrases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrPhrasesRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPatchReviewPhrases_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	phrases := []string{"a phrase"}

	svc.On("AddPhrases", id, phrases).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(map[string][]string{"phrases": phrases})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/reviews/"+id+"/phrases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.AssertExpectations(t)
}

// TestPatchReviewPhrases_ContextBodyLoaded verifies the phrases are read from
// the request body and the correct review ID is forwarded to the service.
func TestPatchReviewPhrases_ContextBodyLoaded(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	review := sampleReview()
	review.ID = bson.NewObjectID()
	phrases := []string{"phrase one", "phrase two"}
	expectedID := review.ID.Hex()

	var capturedID string
	var capturedPhrases []string
	svc.On("AddPhrases", expectedID, phrases).
		Run(func(args mock.Arguments) {
			capturedID = args.String(0)
			capturedPhrases = args.Get(1).([]string)
		}).
		Return(review, nil)

	body, _ := json.Marshal(map[string][]string{"phrases": phrases})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/reviews/"+expectedID+"/phrases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedID, capturedID)
	assert.Equal(t, phrases, capturedPhrases)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteReview
// ---------------------------------------------------------------------------

func TestDeleteReview_Success(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("DeleteByID", id).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	svc.AssertExpectations(t)
}

func TestDeleteReview_InternalError(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("DeleteByID", id).Return(errors.New("db delete failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "db delete failed", resp["message"])

	svc.AssertExpectations(t)
}

func TestDeleteReview_IDPassedToService(t *testing.T) {
	svc := new(mockReviewService)
	h := handler.NewReviewHandler(svc)
	router := setupReviewRouter(h)

	id := bson.NewObjectID().Hex()

	var capturedID string
	svc.On("DeleteByID", id).
		Run(func(args mock.Arguments) { capturedID = args.String(0) }).
		Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, id, capturedID)

	svc.AssertExpectations(t)
}
