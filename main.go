package main

import (
	"example_api/initializers"
	"example_api/repositories"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "example_api/docs"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database
	db, err := initializers.ConnectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize the User repository
	userRepo := repositories.NewUserRepository(db)

	// Set up the router
	r := mux.NewRouter()

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// User routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", userRepo.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", userRepo.GetUserByID).Methods("GET")
	api.HandleFunc("/users/{id}", userRepo.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userRepo.DeleteUser).Methods("DELETE")

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
