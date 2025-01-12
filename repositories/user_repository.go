package repositories

import (
	"context"
	"encoding/json"
	model "example_api/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with email, password, first name, and last name
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.User true "User JSON"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/users [post]
func (repo *UserRepository) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"status":400, "message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" || user.FirstName == "" || user.LastName == "" {
		http.Error(w, `{"status":400, "message":"All fields except ID and JoinDate are required"}`, http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"status":500, "message":"Error hashing password"}`, http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	user.Id = primitive.NewObjectID()
	user.JoinDate = time.Now()

	// Insert into database
	_, err = repo.collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, `{"status":500, "message":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  201,
		"message": fmt.Sprintf("User created successfully with ID: %s", user.Id.Hex()),
		"data":    user,
	})
}

// GetUserByID godoc
// @Summary Get a user by ID
// @Description Retrieve user details by their unique ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/users/{id} [get]
func (repo *UserRepository) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, `{"status":400, "message":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var user model.User
	err = repo.collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		http.Error(w, `{"status":404, "message":"User not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "User retrieved successfully",
		"data":    user,
	})
}

// UpdateUser godoc
// @Summary Update user details
// @Description Update specific fields of a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param updates body map[string]interface{} true "Update fields JSON"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/users/{id} [put]
func (repo *UserRepository) UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, `{"status":400, "message":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"status":400, "message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	allowedFields := map[string]bool{
		"email":     true,
		"firstName": true,
		"lastName":  true,
		"password":  true,
	}

	filteredUpdates := bson.M{}
	for key, value := range updates {
		if allowedFields[key] {
			if key == "password" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%v", value)), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, `{"status":500, "message":"Error hashing password"}`, http.StatusInternalServerError)
					return
				}
				filteredUpdates[key] = string(hashedPassword)
			} else {
				filteredUpdates[key] = value
			}
		}
	}

	if len(filteredUpdates) == 0 {
		http.Error(w, `{"status":400, "message":"No valid fields to update"}`, http.StatusBadRequest)
		return
	}

	_, err = repo.collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": filteredUpdates})
	if err != nil {
		http.Error(w, `{"status":500, "message":"Failed to update user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "User updated successfully",
	})
}

// DeleteUser godoc
// @Summary Delete a user by ID
// @Description Remove a user from the database using their unique ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/users/{id} [delete]
func (repo *UserRepository) DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, `{"status":400, "message":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	_, err = repo.collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, `{"status":500, "message":"Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "User deleted successfully",
	})
}
