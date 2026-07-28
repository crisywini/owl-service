package model

import "go.mongodb.org/mongo-driver/v2/bson"

type Book struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string        `bson:"title" json:"title"`
	Authors       []string      `bson:"authors" json:"authors"`
	Publisher     string        `bson:"publisher" json:"publisher"`
	PublishedYear int           `bson:"published_year" json:"published_year"`
	ISBN10        string        `bson:"isbn10" json:"isbn10"`
	ISBN13        string        `bson:"isbn13" json:"isbn13"`
	Description   string        `bson:"description" json:"description"`
	Genre         []string      `bson:"genre" json:"genre"`
}

type BookBuilder struct {
	book Book
}

func NewBookBuilder() *BookBuilder {
	return &BookBuilder{book: Book{}}
}

func (b *BookBuilder) WithTitle(title string) *BookBuilder {
	b.book.Title = title
	return b
}

func (b *BookBuilder) WithAuthors(authors []string) *BookBuilder {
	b.book.Authors = authors
	return b
}

func (b *BookBuilder) WithPublisher(publisher string) *BookBuilder {
	b.book.Publisher = publisher
	return b
}

func (b *BookBuilder) WithPublishedYear(year int) *BookBuilder {
	b.book.PublishedYear = year
	return b
}

func (b *BookBuilder) WithISBN10(isbn10 string) *BookBuilder {
	b.book.ISBN10 = isbn10
	return b
}

func (b *BookBuilder) WithISBN13(isbn13 string) *BookBuilder {
	b.book.ISBN13 = isbn13
	return b
}

func (b *BookBuilder) WithDescription(description string) *BookBuilder {
	b.book.Description = description
	return b
}

func (b *BookBuilder) WithGenre(genre []string) *BookBuilder {
	b.book.Genre = genre
	return b
}

func (b *BookBuilder) Build() Book {
	return b.book
}
