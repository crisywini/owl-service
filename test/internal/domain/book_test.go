package domain_test

import (
	"testing"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBookWithCorrectParameters_Success(t *testing.T) {
	bookBuilder := domain.NewBookBuilder()
	titleExpected := "Giovanni's Room"
	bookBuilder.WithTitle(titleExpected)
	descriptionExpected := "Giovanni's Room is a 1956 novel by James Baldwin. The book concerns the events in the life of an American man living in Paris and his feelings and frustrations with his relationships with other men, particularly an Italian bartender named Giovanni whom he meets at a Parisian gay bar."
	bookBuilder.WithDescription(descriptionExpected)
	genresExpected := []string{"Artists"}
	bookBuilder.WithGenre(genresExpected)
	ISBN10Expected := uuid.NewString()
	bookBuilder.WithISBN10(ISBN10Expected)
	ISBN13Expected := uuid.NewString()
	bookBuilder.WithISBN13(ISBN13Expected)
	publisherExpected := "The Dial Press"
	bookBuilder.WithPublisher(publisherExpected)
	publishedYearExpected := 1956
	bookBuilder.WithPublishedYear(publishedYearExpected)

	book := bookBuilder.Build()

	assert.Equal(t, titleExpected, book.Title)
	assert.Equal(t, descriptionExpected, book.Description)
	assert.Equal(t, genresExpected, book.Genre)
	assert.Equal(t, ISBN10Expected, book.ISBN10)
	assert.Equal(t, ISBN13Expected, book.ISBN13)
	assert.Equal(t, publisherExpected, book.Publisher)
	assert.Equal(t, publishedYearExpected, book.PublishedYear)

}
