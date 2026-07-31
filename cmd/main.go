package main

import (
	"context"
	"log"
	"os"

	"github.com/crisywini/owl-service/internal/handler"
	"github.com/crisywini/owl-service/internal/repository"
	"github.com/crisywini/owl-service/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found 🧐, falling back to system env 🙈")
	}

	r := gin.Default()
	r.Use(cors.New(
		cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true}))

	var databaseUri string

	if databaseUri = os.Getenv("MONGODB_URI"); databaseUri == "" {
		log.Fatal("You must set the MongoDB URI! 😖")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(databaseUri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	database := client.Database("owl-service")
	bookRepository := repository.NewBookRepository(database)
	bookService := usecase.NewBookService(bookRepository)

	reviewRepository := repository.NewReviewRepository(database)
	reviewService := usecase.NewReviewService(reviewRepository, bookRepository)

	bookHandler := handler.NewBookHandler(bookService)
	reviewHandler := handler.NewReviewHandler(reviewService)

	r.POST("/books", bookHandler.PostBook)
	r.GET("/books/:id", bookHandler.GetBook)
	r.GET("/books", bookHandler.GetBooks)
	r.PUT("/books/:id", bookHandler.PutBook)
	r.DELETE("/books/:id", bookHandler.DeleteBook)

	r.POST("/reviews", reviewHandler.PostReview)
	r.GET("/reviews/:id", reviewHandler.GetReview)
	r.GET("/reviews", reviewHandler.GetReviews)
	r.PUT("/reviews/:id", reviewHandler.PutReview)
	r.DELETE("/reviews/:id", reviewHandler.DeleteReview)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
