package main

import (
	"encoding/json"
	"net/http"

	"github.com/SirLouen/chirpy-bootdev/src/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		ID          uuid.UUID `json:"id"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	request := struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No key found", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid", nil)
		return
	}

	userID, err := uuid.Parse(request.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if request.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Unsupported event type", nil)
		return
	}

	err = cfg.db.MarkUserChirpsAsRed(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark user's chirps as red", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, Response{
		ID:          userID,
		IsChirpyRed: true,
	})
}
