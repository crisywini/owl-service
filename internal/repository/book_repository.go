package repository

import (
	"context"
	"time"

	"github.com/crisywini/owl-service/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BookRepository struct {
	collection *mongo.Collection
}

func NewBookRepository(db *mongo.Database) *BookRepository {
	return &BookRepository{
		collection: db.Collection("books"),
	}
}

func (r *BookRepository) Save(book *model.Book) (*model.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	book.ID = bson.NewObjectID()

	_, err := r.collection.InsertOne(ctx, book)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (r *BookRepository) FindAll() ([]model.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var books []model.Book
	if err := cursor.All(ctx, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *BookRepository) FindByID(id string) (*model.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var book model.Book
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) Update(id string, updated *model.Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"title":          updated.Title,
			"authors":        updated.Authors,
			"publisher":      updated.Publisher,
			"published_year": updated.PublishedYear,
			"isbn10":         updated.ISBN10,
			"isbn13":         updated.ISBN13,
			"description":    updated.Description,
			"genre":          updated.Genre,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *BookRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
