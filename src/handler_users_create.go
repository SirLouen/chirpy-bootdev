package main

import (
	"encoding/json"
	"net/http"

	"github.com/SirLouen/chirpy-bootdev/src/auth"
	"github.com/SirLouen/chirpy-bootdev/src/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type request User
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	password, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: password,
	}

	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}
