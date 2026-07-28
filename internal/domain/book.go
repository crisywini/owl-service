package domain

type Book struct {
	ID            string
	Title         string
	Authors       []string
	Publisher     string
	PublishedYear int
	ISBN10        string
	ISBN13        string
	Description   string
	Genre         []string
}

type BookBuilder struct {
	book Book
}

func NewBookBuilder() *BookBuilder {
	return &BookBuilder{
		book: Book{},
	}
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

func (b *BookBuilder) WithISBN10(ISBN10 string) *BookBuilder {
	b.book.ISBN10 = ISBN10
	return b
}

func (b *BookBuilder) WithISBN13(ISBN13 string) *BookBuilder {
	b.book.ISBN13 = ISBN13
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

func (b *BookBuilder) WithPublishedYear(year int) *BookBuilder {
	b.book.PublishedYear = year
	return b
}

func (b *BookBuilder) Build() Book {
	return b.book
}
