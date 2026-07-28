package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Review struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	BookID          bson.ObjectID `bson:"book_id" json:"book_id"`
	Rate            int           `bson:"rate" json:"rate"`
	Description     string        `bson:"description" json:"description"`
	StartDate       time.Time     `bson:"start_date" json:"start_date"`
	FinishDate      time.Time     `bson:"finish_date" json:"finish_date"`
	FavoritePhrases []string      `bson:"favorite_phrases" json:"favorite_phrases"`
}

type ReviewBuilder struct {
	review Review
}

func NewReviewBuilder(bookID bson.ObjectID) *ReviewBuilder {
	return &ReviewBuilder{review: Review{BookID: bookID}}
}

func (b *ReviewBuilder) WithRate(rate int) *ReviewBuilder {
	b.review.Rate = rate
	return b
}

func (b *ReviewBuilder) WithDescription(description string) *ReviewBuilder {
	b.review.Description = description
	return b
}

func (b *ReviewBuilder) WithStartDate(date time.Time) *ReviewBuilder {
	b.review.StartDate = date
	return b
}

func (b *ReviewBuilder) WithFinishDate(date time.Time) *ReviewBuilder {
	b.review.FinishDate = date
	return b
}

func (b *ReviewBuilder) WithFavoritePhrases(phrases []string) *ReviewBuilder {
	b.review.FavoritePhrases = phrases
	return b
}

func (b *ReviewBuilder) AddFavoritePhrase(phrase string) *ReviewBuilder {
	b.review.FavoritePhrases = append(b.review.FavoritePhrases, phrase)
	return b
}

func (b *ReviewBuilder) Build() Review {
	return b.review
}
