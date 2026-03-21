package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SirLouen/chirpy-bootdev/src/auth"
	"github.com/SirLouen/chirpy-bootdev/src/database"
	"github.com/google/uuid"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// validateRefreshToken extracts and validates the refresh token from the request.
func (cfg *apiConfig) validateRefreshToken(r *http.Request) (database.RefreshToken, error) {
	bearerToken, err := auth.GetBearerToken(r)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("missing refresh token: %w", err)
	}

	tokenUUID, err := uuid.Parse(bearerToken)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), tokenUUID)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("refresh token not found: %w", err)
	}

	if refreshToken.ExpiresAt.Before(time.Now().UTC()) || refreshToken.RevokedAt.Valid {
		return database.RefreshToken{}, fmt.Errorf("expired or revoked refresh token")
	}

	return refreshToken, nil
}
