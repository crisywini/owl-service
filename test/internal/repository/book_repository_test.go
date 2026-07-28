package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/crisywini/owl-service/internal/domain"
	"github.com/crisywini/owl-service/internal/repository"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func setupBookRepo(t *testing.T) (*repository.BookRepository, func()) {
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

	repo := repository.NewBookRepository(client.Database("owl_test"))

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(shutdownCtx)
		_ = container.Terminate(ctx)
	}

	return repo, cleanup
}

func TestBookRepository_SaveAndRetrieve(t *testing.T) {
	repo, cleanup := setupBookRepo(t)
	defer cleanup()

	book := domain.NewBookBuilder().
		WithTitle("Giovanni's Room").
		WithAuthors([]string{"James Baldwin"}).
		WithPublisher("The Dial Press").
		WithPublishedYear(1956).
		WithISBN10("0345806565").
		WithISBN13("9780345806567").
		WithDescription("A novel about love and identity in Paris.").
		WithGenre([]string{"Fiction", "LGBTQ+"}).
		Build()

	saved, err := repo.Save(&book)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if saved.ID.IsZero() {
		t.Fatal("expected ID to be set after Save, got zero value")
	}
	if saved.Title != "Giovanni's Room" {
		t.Errorf("Title = %q, want %q", saved.Title, "Giovanni's Room")
	}
	if len(saved.Authors) != 1 || saved.Authors[0] != "James Baldwin" {
		t.Errorf("Authors = %v, want [James Baldwin]", saved.Authors)
	}
	if saved.Publisher != "The Dial Press" {
		t.Errorf("Publisher = %q, want %q", saved.Publisher, "The Dial Press")
	}
	if saved.PublishedYear != 1956 {
		t.Errorf("PublishedYear = %d, want 1956", saved.PublishedYear)
	}
}

func TestBookRepository_FindAll(t *testing.T) {
	repo, cleanup := setupBookRepo(t)
	defer cleanup()

	seeds := []domain.Book{
		domain.NewBookBuilder().WithTitle("Giovanni's Room").WithAuthors([]string{"James Baldwin"}).Build(),
		domain.NewBookBuilder().WithTitle("1984").WithAuthors([]string{"George Orwell"}).Build(),
		domain.NewBookBuilder().WithTitle("Dune").WithAuthors([]string{"Frank Herbert"}).Build(),
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
		t.Fatalf("FindAll() returned %d books, want %d", len(found), len(seeds))
	}

	titleSet := make(map[string]bool, len(found))
	for _, b := range found {
		titleSet[b.Title] = true
	}
	for _, seed := range seeds {
		if !titleSet[seed.Title] {
			t.Errorf("FindAll() missing book %q", seed.Title)
		}
	}
}

func TestBookRepository_FindByID(t *testing.T) {
	repo, cleanup := setupBookRepo(t)
	defer cleanup()

	book := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		WithPublisher("Chilton Books").
		WithPublishedYear(1965).
		Build()

	saved, err := repo.Save(&book)
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
		if found.Title != saved.Title {
			t.Errorf("Title = %q, want %q", found.Title, saved.Title)
		}
		if found.PublishedYear != saved.PublishedYear {
			t.Errorf("PublishedYear = %d, want %d", found.PublishedYear, saved.PublishedYear)
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

func TestBookRepository_Update(t *testing.T) {
	repo, cleanup := setupBookRepo(t)
	defer cleanup()

	book := domain.NewBookBuilder().
		WithTitle("Dune").
		WithAuthors([]string{"Frank Herbert"}).
		WithPublisher("Chilton Books").
		WithPublishedYear(1965).
		Build()

	saved, err := repo.Save(&book)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Run("updates fields and persists", func(t *testing.T) {
		updated := domain.NewBookBuilder().
			WithTitle("Dune Messiah").
			WithAuthors([]string{"Frank Herbert"}).
			WithPublisher("Putnam").
			WithPublishedYear(1969).
			WithGenre([]string{"Science Fiction"}).
			Build()

		if err := repo.Update(saved.ID.Hex(), &updated); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		found, err := repo.FindByID(saved.ID.Hex())
		if err != nil {
			t.Fatalf("FindByID() after Update error = %v", err)
		}
		if found.Title != updated.Title {
			t.Errorf("Title = %q, want %q", found.Title, updated.Title)
		}
		if found.Publisher != updated.Publisher {
			t.Errorf("Publisher = %q, want %q", found.Publisher, updated.Publisher)
		}
		if found.PublishedYear != updated.PublishedYear {
			t.Errorf("PublishedYear = %d, want %d", found.PublishedYear, updated.PublishedYear)
		}
		if found.ID != saved.ID {
			t.Errorf("ID changed after update: got %v, want %v", found.ID, saved.ID)
		}
	})

	t.Run("unknown id returns no error", func(t *testing.T) {
		patch := domain.NewBookBuilder().WithTitle("Ghost Book").Build()
		if err := repo.Update("000000000000000000000000", &patch); err != nil {
			t.Errorf("Update() unexpected error for unknown ID: %v", err)
		}
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		patch := domain.NewBookBuilder().WithTitle("Ghost Book").Build()
		if err := repo.Update("not-a-valid-id", &patch); err == nil {
			t.Fatal("Update() expected error for invalid ID, got nil")
		}
	})
}

func TestBookRepository_Delete(t *testing.T) {
	repo, cleanup := setupBookRepo(t)
	defer cleanup()

	book := domain.NewBookBuilder().
		WithTitle("1984").
		WithAuthors([]string{"George Orwell"}).
		WithPublishedYear(1949).
		Build()

	saved, err := repo.Save(&book)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Run("deletes existing book", func(t *testing.T) {
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
