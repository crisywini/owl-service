package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/crisywini/owl-service/internal/model"
	"github.com/crisywini/owl-service/internal/repository"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func setupReviewRepo(t *testing.T) (*repository.ReviewRepository, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("failed to start mongodb container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to connect to mongodb: %v", err)
	}

	repo := repository.NewReviewRepository(client.Database("owl_test"))

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(shutdownCtx)
		_ = container.Terminate(ctx)
	}

	return repo, cleanup
}

func TestReviewRepository_SaveAndRetrieve(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookID := bson.NewObjectID()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	finish := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	review := model.NewReviewBuilder(bookID).
		WithRate(5).
		WithDescription("An absolute masterpiece.").
		WithStartDate(start).
		WithFinishDate(finish).
		AddFavoritePhrase("The heart is deceitful above all things.").
		Build()

	saved, err := repo.Save(&review)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if saved.ID.IsZero() {
		t.Fatal("expected ID to be set after Save, got zero value")
	}
	if saved.Rate != 5 {
		t.Errorf("Rate = %d, want 5", saved.Rate)
	}
	if saved.Description != "An absolute masterpiece." {
		t.Errorf("Description = %q, want %q", saved.Description, "An absolute masterpiece.")
	}
	if saved.BookID != bookID {
		t.Errorf("BookID = %v, want %v", saved.BookID, bookID)
	}
	if len(saved.FavoritePhrases) != 1 {
		t.Errorf("FavoritePhrases count = %d, want 1", len(saved.FavoritePhrases))
	}
}

func TestReviewRepository_FindAll(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookID := bson.NewObjectID()
	seeds := []model.Review{
		model.NewReviewBuilder(bookID).WithRate(5).WithDescription("Loved it!").Build(),
		model.NewReviewBuilder(bookID).WithRate(4).WithDescription("Very good.").Build(),
		model.NewReviewBuilder(bson.NewObjectID()).WithRate(3).WithDescription("It was okay.").Build(),
	}

	for i := range seeds {
		if _, err := repo.Save(&seeds[i]); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	found, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(found) != len(seeds) {
		t.Fatalf("FindAll() returned %d reviews, want %d", len(found), len(seeds))
	}

	descSet := make(map[string]bool, len(found))
	for _, r := range found {
		descSet[r.Description] = true
	}
	for _, seed := range seeds {
		if !descSet[seed.Description] {
			t.Errorf("FindAll() missing review %q", seed.Description)
		}
	}
}

func TestReviewRepository_FindByID(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookID := bson.NewObjectID()
	review := model.NewReviewBuilder(bookID).
		WithRate(4).
		WithDescription("Really enjoyed this one.").
		AddFavoritePhrase("It was the best of times.").
		Build()

	saved, err := repo.Save(&review)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(saved.ID.Hex())
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if found.ID != saved.ID {
			t.Errorf("ID = %v, want %v", found.ID, saved.ID)
		}
		if found.Rate != saved.Rate {
			t.Errorf("Rate = %d, want %d", found.Rate, saved.Rate)
		}
		if found.Description != saved.Description {
			t.Errorf("Description = %q, want %q", found.Description, saved.Description)
		}
		if len(found.FavoritePhrases) != len(saved.FavoritePhrases) {
			t.Errorf("FavoritePhrases count = %d, want %d", len(found.FavoritePhrases), len(saved.FavoritePhrases))
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID("000000000000000000000000")
		if err == nil {
			t.Fatal("FindByID() expected error for unknown ID, got nil")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := repo.FindByID("not-a-valid-id")
		if err == nil {
			t.Fatal("FindByID() expected error for invalid ID, got nil")
		}
	})
}

func TestReviewRepository_FindByBookID(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookA := bson.NewObjectID()
	bookB := bson.NewObjectID()

	reviews := []model.Review{
		model.NewReviewBuilder(bookA).WithRate(5).WithDescription("Great book A!").Build(),
		model.NewReviewBuilder(bookA).WithRate(4).WithDescription("Really liked book A.").Build(),
		model.NewReviewBuilder(bookB).WithRate(3).WithDescription("Book B was average.").Build(),
	}

	for i := range reviews {
		if _, err := repo.Save(&reviews[i]); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	t.Run("returns reviews for a specific book", func(t *testing.T) {
		found, err := repo.FindByBookID(bookA.Hex())
		if err != nil {
			t.Fatalf("FindByBookID() error = %v", err)
		}
		if len(found) != 2 {
			t.Fatalf("FindByBookID() returned %d reviews, want 2", len(found))
		}
		for _, r := range found {
			if r.BookID != bookA {
				t.Errorf("FindByBookID() returned review with wrong BookID: %v", r.BookID)
			}
		}
	})

	t.Run("returns empty slice for book with no reviews", func(t *testing.T) {
		found, err := repo.FindByBookID(bson.NewObjectID().Hex())
		if err != nil {
			t.Fatalf("FindByBookID() error = %v", err)
		}
		if len(found) != 0 {
			t.Fatalf("FindByBookID() returned %d reviews, want 0", len(found))
		}
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		_, err := repo.FindByBookID("not-a-valid-id")
		if err == nil {
			t.Fatal("FindByBookID() expected error for invalid ID, got nil")
		}
	})
}

func TestReviewRepository_Update(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookID := bson.NewObjectID()
	review := model.NewReviewBuilder(bookID).
		WithRate(3).
		WithDescription("It was okay.").
		Build()

	saved, err := repo.Save(&review)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Run("updates fields and persists", func(t *testing.T) {
		updated := model.NewReviewBuilder(bookID).
			WithRate(5).
			WithDescription("Changed my mind, it was amazing!").
			AddFavoritePhrase("A memorable quote.").
			Build()

		if err := repo.Update(saved.ID.Hex(), &updated); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		found, err := repo.FindByID(saved.ID.Hex())
		if err != nil {
			t.Fatalf("FindByID() after Update error = %v", err)
		}
		if found.Rate != updated.Rate {
			t.Errorf("Rate = %d, want %d", found.Rate, updated.Rate)
		}
		if found.Description != updated.Description {
			t.Errorf("Description = %q, want %q", found.Description, updated.Description)
		}
		if len(found.FavoritePhrases) != 1 {
			t.Errorf("FavoritePhrases count = %d, want 1", len(found.FavoritePhrases))
		}
		if found.ID != saved.ID {
			t.Errorf("ID changed after update: got %v, want %v", found.ID, saved.ID)
		}
	})

	t.Run("unknown id returns no error", func(t *testing.T) {
		patch := model.NewReviewBuilder(bookID).WithRate(1).Build()
		if err := repo.Update("000000000000000000000000", &patch); err != nil {
			t.Errorf("Update() unexpected error for unknown ID: %v", err)
		}
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		patch := model.NewReviewBuilder(bookID).WithRate(1).Build()
		if err := repo.Update("not-a-valid-id", &patch); err == nil {
			t.Fatal("Update() expected error for invalid ID, got nil")
		}
	})
}

func TestReviewRepository_Delete(t *testing.T) {
	repo, cleanup := setupReviewRepo(t)
	defer cleanup()

	bookID := bson.NewObjectID()
	review := model.NewReviewBuilder(bookID).
		WithRate(4).
		WithDescription("A solid read.").
		Build()

	saved, err := repo.Save(&review)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Run("deletes existing review", func(t *testing.T) {
		if err := repo.Delete(saved.ID.Hex()); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := repo.FindByID(saved.ID.Hex())
		if err == nil {
			t.Fatal("FindByID() expected error after Delete, got nil")
		}
	})

	t.Run("unknown id returns no error", func(t *testing.T) {
		if err := repo.Delete("000000000000000000000000"); err != nil {
			t.Errorf("Delete() unexpected error for unknown ID: %v", err)
		}
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		if err := repo.Delete("not-a-valid-id"); err == nil {
			t.Fatal("Delete() expected error for invalid ID, got nil")
		}
	})
}
