package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SirLouen/chirpy-bootdev/src/auth"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Expire   int64  `json:"expire_in_seconds,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req parameters
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	if req.Expire == 0 || req.Expire > 3600 {
		req.Expire = 3600
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(req.Expire)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
