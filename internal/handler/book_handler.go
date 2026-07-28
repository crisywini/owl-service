package handler

import (
	"errors"
	"net/http"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	service usecase.BookServicePort
}

func NewBookHandler(service usecase.BookServicePort) *BookHandler {
	return &BookHandler{service: service}
}

func (b *BookHandler) PostBook(c *gin.Context) {
	var book domain.Book

	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed Book",
			"error":   err.Error(),
		})
		return
	}

	response, err := b.service.Create(&book)
	if err != nil {
		if errors.Is(err, usecase.ErrBookTitleRequired) || errors.Is(err, usecase.ErrBookAuthorsRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": response.ID.Hex()})
}

func (b *BookHandler) GetBook(c *gin.Context) {
	id := c.Param("id")

	book, err := b.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func (b *BookHandler) GetBooks(c *gin.Context) {
	books, err := b.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, books)
}

func (b *BookHandler) PutBook(c *gin.Context) {
	id := c.Param("id")

	var book domain.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed Book",
			"error":   err.Error(),
		})
		return
	}

	updated, err := b.service.Update(id, &book)
	if err != nil {
		if errors.Is(err, usecase.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		if errors.Is(err, usecase.ErrBookTitleRequired) || errors.Is(err, usecase.ErrAuthorsCannotBeModified) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (b *BookHandler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	if err := b.service.DeleteByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
