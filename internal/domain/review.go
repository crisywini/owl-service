package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID              string
	Rate            int
	Description     string
	StartDate       time.Time
	FinishDate      time.Time
	FavoritePhrases []string
}

func NewReview(rate int, description string, startDate, finishDate time.Time, favoritePhrases []string) *Review {

	return &Review{
		ID:              uuid.NewString(),
		Rate:            rate,
		Description:     description,
		StartDate:       startDate,
		FinishDate:      finishDate,
		FavoritePhrases: favoritePhrases,
	}
}
