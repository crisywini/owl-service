package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

// Compile-time assertion: *mockBookService must implement usecase.BookServicePort.
var _ usecase.BookServicePort = (*mockBookService)(nil)

type mockBookService struct {
	mock.Mock
}

func (m *mockBookService) Create(book *domain.Book) (*domain.Book, error) {
	args := m.Called(book)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *mockBookService) GetByID(id string) (*domain.Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *mockBookService) GetAll() ([]domain.Book, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Book), args.Error(1)
}

func (m *mockBookService) Update(id string, updates *domain.Book) (*domain.Book, error) {
	args := m.Called(id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *mockBookService) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupRouter(h *handler.BookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/books", h.PostBook)
	r.GET("/books", h.GetBooks)
	r.GET("/books/:id", h.GetBook)
	r.PUT("/books/:id", h.PutBook)
	r.DELETE("/books/:id", h.DeleteBook)
	return r
}

func sampleBook() *domain.Book {
	book := domain.NewBookBuilder().
		WithTitle("The Go Programming Language").
		WithAuthors([]string{"Alan Donovan", "Brian Kernighan"}).
		WithPublisher("Addison-Wesley").
		WithPublishedYear(2015).
		WithISBN10("0134190440").
		WithISBN13("978-0134190440").
		WithDescription("A comprehensive guide to Go.").
		WithGenre([]string{"Programming", "Technology"}).
		Build()
	return &book
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewBuffer(b)
}

// ---------------------------------------------------------------------------
// PostBook
// ---------------------------------------------------------------------------

func TestPostBook_Success(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	savedBook := *book
	savedBook.ID = bson.NewObjectID()

	svc.On("Create", mock.AnythingOfType("*domain.Book")).Return(&savedBook, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, savedBook.ID.Hex(), resp["id"])

	svc.AssertExpectations(t)
}

func TestPostBook_MalformedJSON(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Malformed Book", resp["message"])

	svc.AssertNotCalled(t, "Create")
}

func TestPostBook_TitleRequired(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.Title = ""

	svc.On("Create", mock.AnythingOfType("*domain.Book")).Return(nil, usecase.ErrBookTitleRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrBookTitleRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPostBook_AuthorsRequired(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.Authors = []string{}

	svc.On("Create", mock.AnythingOfType("*domain.Book")).Return(nil, usecase.ErrBookAuthorsRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrBookAuthorsRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPostBook_InternalError(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	svc.On("Create", mock.AnythingOfType("*domain.Book")).Return(nil, errors.New("db connection failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "db connection failed", resp["message"])

	svc.AssertExpectations(t)
}

// TestPostBook_ContextBodyLoaded verifies that the gin context properly reads
// the JSON body and passes a populated Book pointer to the service.
func TestPostBook_ContextBodyLoaded(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	savedBook := *book
	savedBook.ID = bson.NewObjectID()

	var capturedBook *domain.Book
	svc.On("Create", mock.AnythingOfType("*domain.Book")).
		Run(func(args mock.Arguments) {
			capturedBook = args.Get(0).(*domain.Book)
		}).
		Return(&savedBook, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/books", jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotNil(t, capturedBook)
	assert.Equal(t, book.Title, capturedBook.Title)
	assert.Equal(t, book.Authors, capturedBook.Authors)
	assert.Equal(t, book.Publisher, capturedBook.Publisher)
	assert.Equal(t, book.ISBN10, capturedBook.ISBN10)
	assert.Equal(t, book.ISBN13, capturedBook.ISBN13)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetBook
// ---------------------------------------------------------------------------

func TestGetBook_Success(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.ID = bson.NewObjectID()

	svc.On("GetByID", book.ID.Hex()).Return(book, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books/"+book.ID.Hex(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.Book
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, book.Title, resp.Title)
	assert.Equal(t, book.Authors, resp.Authors)

	svc.AssertExpectations(t)
}

func TestGetBook_NotFound(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("GetByID", id).Return(nil, usecase.ErrBookNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrBookNotFound.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestGetBook_IDPassedToService(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.ID = bson.NewObjectID()
	expectedID := book.ID.Hex()

	var capturedID string
	svc.On("GetByID", expectedID).
		Run(func(args mock.Arguments) {
			capturedID = args.String(0)
		}).
		Return(book, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books/"+expectedID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedID, capturedID)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetBooks
// ---------------------------------------------------------------------------

func TestGetBooks_ReturnsList(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	b1 := sampleBook()
	b1.ID = bson.NewObjectID()
	b2 := sampleBook()
	b2.Title = "Clean Code"
	b2.ID = bson.NewObjectID()

	svc.On("GetAll").Return([]domain.Book{*b1, *b2}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Book
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)

	svc.AssertExpectations(t)
}

func TestGetBooks_EmptyList(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	svc.On("GetAll").Return([]domain.Book{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Book
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)

	svc.AssertExpectations(t)
}

func TestGetBooks_InternalError(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	svc.On("GetAll").Return(nil, errors.New("db timeout"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "db timeout", resp["message"])

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// PutBook
// ---------------------------------------------------------------------------

func TestPutBook_Success(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.ID = bson.NewObjectID()
	updatedBook := *book
	updatedBook.Description = "Updated description."

	svc.On("Update", book.ID.Hex(), mock.AnythingOfType("*domain.Book")).Return(&updatedBook, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+book.ID.Hex(), jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.Book
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, updatedBook.Description, resp.Description)

	svc.AssertExpectations(t)
}

func TestPutBook_MalformedJSON(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+id, bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Malformed Book", resp["message"])

	svc.AssertNotCalled(t, "Update")
}

func TestPutBook_NotFound(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	book := sampleBook()

	svc.On("Update", id, mock.AnythingOfType("*domain.Book")).Return(nil, usecase.ErrBookNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+id, jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrBookNotFound.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutBook_TitleRequired(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	book := sampleBook()
	book.Title = ""

	svc.On("Update", id, mock.AnythingOfType("*domain.Book")).Return(nil, usecase.ErrBookTitleRequired)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+id, jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrBookTitleRequired.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutBook_AuthorsCannotBeModified(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	book := sampleBook()
	book.Authors = []string{"New Author"}

	svc.On("Update", id, mock.AnythingOfType("*domain.Book")).Return(nil, usecase.ErrAuthorsCannotBeModified)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+id, jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, usecase.ErrAuthorsCannotBeModified.Error(), resp["message"])

	svc.AssertExpectations(t)
}

func TestPutBook_InternalError(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	book := sampleBook()

	svc.On("Update", id, mock.AnythingOfType("*domain.Book")).Return(nil, errors.New("db write failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+id, jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.AssertExpectations(t)
}

// TestPutBook_ContextBodyLoaded verifies the request body is fully deserialized
// and the correct ID is extracted from the URL parameter.
func TestPutBook_ContextBodyLoaded(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	book := sampleBook()
	book.ID = bson.NewObjectID()
	expectedID := book.ID.Hex()

	var capturedID string
	var capturedBook *domain.Book

	svc.On("Update", expectedID, mock.AnythingOfType("*domain.Book")).
		Run(func(args mock.Arguments) {
			capturedID = args.String(0)
			capturedBook = args.Get(1).(*domain.Book)
		}).
		Return(book, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/books/"+expectedID, jsonBody(t, book))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedID, capturedID)
	assert.NotNil(t, capturedBook)
	assert.Equal(t, book.Title, capturedBook.Title)
	assert.Equal(t, book.Publisher, capturedBook.Publisher)

	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteBook
// ---------------------------------------------------------------------------

func TestDeleteBook_Success(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("DeleteByID", id).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/books/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	svc.AssertExpectations(t)
}

func TestDeleteBook_InternalError(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()
	svc.On("DeleteByID", id).Return(errors.New("db delete failed"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/books/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "db delete failed", resp["message"])

	svc.AssertExpectations(t)
}

// TestDeleteBook_IDPassedToService verifies the path parameter is correctly
// extracted and forwarded to the service.
func TestDeleteBook_IDPassedToService(t *testing.T) {
	svc := new(mockBookService)
	h := handler.NewBookHandler(svc)
	router := setupRouter(h)

	id := bson.NewObjectID().Hex()

	var capturedID string
	svc.On("DeleteByID", id).
		Run(func(args mock.Arguments) {
			capturedID = args.String(0)
		}).
		Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/books/"+id, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, id, capturedID)

	svc.AssertExpectations(t)
}
