package main

import (
	"log"
	"net/http"

	"example.com/web/handlers"
	"example.com/web/middleware"
)

func main() {
	router := http.NewServeMux()

	// Add middleware
	auth := middleware.NewAuthMiddleware()
	logger := middleware.NewLoggingMiddleware()

	// Register routes
	router.Handle("/users", auth(handlers.HandleUsers()))
	router.Handle("/posts", auth(handlers.HandlePosts()))
	router.Handle("/health", logger(handlers.HandleHealth()))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
