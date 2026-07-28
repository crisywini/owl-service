package handler

import (
	"errors"
	"net/http"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service usecase.ReviewServicePort
}

func NewReviewHandler(service usecase.ReviewServicePort) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (r *ReviewHandler) PostReview(c *gin.Context) {
	var review domain.Review

	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed Review",
			"error":   err.Error(),
		})
		return
	}

	response, err := r.service.Create(&review)
	if err != nil {
		if errors.Is(err, usecase.ErrReviewBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		if errors.Is(err, usecase.ErrReviewRateRequired) || errors.Is(err, usecase.ErrReviewDescriptionRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": response.ID.Hex()})
}

func (r *ReviewHandler) GetReview(c *gin.Context) {
	id := c.Param("id")

	review, err := r.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

func (r *ReviewHandler) GetReviews(c *gin.Context) {
	reviews, err := r.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (r *ReviewHandler) GetReviewsByBook(c *gin.Context) {
	bookID := c.Param("bookId")

	reviews, err := r.service.GetByBookID(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (r *ReviewHandler) PutReview(c *gin.Context) {
	id := c.Param("id")

	var review domain.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed Review",
			"error":   err.Error(),
		})
		return
	}

	updated, err := r.service.Update(id, &review)
	if err != nil {
		if errors.Is(err, usecase.ErrReviewBookIDImmutable) ||
			errors.Is(err, usecase.ErrReviewRateRequired) ||
			errors.Is(err, usecase.ErrReviewDescriptionRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

type addPhrasesRequest struct {
	Phrases []string `json:"phrases"`
}

func (r *ReviewHandler) PatchReviewPhrases(c *gin.Context) {
	id := c.Param("id")

	var req addPhrasesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed request",
			"error":   err.Error(),
		})
		return
	}

	updated, err := r.service.AddPhrases(id, req.Phrases)
	if err != nil {
		if errors.Is(err, usecase.ErrPhrasesRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (r *ReviewHandler) DeleteReview(c *gin.Context) {
	id := c.Param("id")

	if err := r.service.DeleteByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
