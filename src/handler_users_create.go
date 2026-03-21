package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SirLouen/chirpy-bootdev/src/auth"
	"github.com/SirLouen/chirpy-bootdev/src/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type request User
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Email is required"}`))
		return
	}

	password, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to hash password"}`))
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: password,
	}

	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to create user"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
