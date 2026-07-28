package repository

import (
	"context"
	"time"

	"github.com/crisywini/owl-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ReviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) *ReviewRepository {
	return &ReviewRepository{
		collection: db.Collection("reviews"),
	}
}

func (r *ReviewRepository) Save(review *domain.Review) (*domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	review.ID = bson.NewObjectID()

	_, err := r.collection.InsertOne(ctx, review)
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *ReviewRepository) FindAll() ([]domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []domain.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) FindByID(id string) (*domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var review domain.Review
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) FindByBookID(bookID string) ([]domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"book_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []domain.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) Update(id string, updated *domain.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"rate":             updated.Rate,
			"description":      updated.Description,
			"start_date":       updated.StartDate,
			"finish_date":      updated.FinishDate,
			"favorite_phrases": updated.FavoritePhrases,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *ReviewRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
