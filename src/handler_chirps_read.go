package main

import (
	"net/http"

	"github.com/SirLouen/chirpy-bootdev/src/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpRead(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRead(w http.ResponseWriter, r *http.Request) {

	authorIDStr := r.URL.Query().Get("author_id")

	var chirps []database.Chirp
	var err error
	if authorIDStr != "" {
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", parseErr)
			return
		}
		chirps, err = cfg.db.GetChirpsByAuthor(r.Context(), authorID)
	} else {
		chirps, err = cfg.db.GetAllChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps", err)
		return
	}

	response := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		response[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}
