package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanProfanity(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Failed to get chirps: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to get chirps"}`))
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (cfg *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	type request Chirp
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}
	if req.Body == "" || req.UserID == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Body and UserID are required"}`))
		return
	}

	cleanBody := cleanProfanity(req.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody,
		UserID: req.UserID,
	})
	if err != nil {
		log.Printf("Failed to create chirp: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to create chirp"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
	})
}

func (cfg *apiConfig) postUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to create user"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email,
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	port := "8080"
	filepathRoot := "./"
	apiRoute := "/api/"
	adminRoute := "/admin/"
	apiCfg := apiConfig{db: dbQueries, platform: os.Getenv("PLATFORM")}
	mux := http.NewServeMux()

	// Static file server for /app/ route
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))

	// API Routes
	mux.HandleFunc("GET "+apiRoute+"healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST "+apiRoute+"users", apiCfg.postUserHandler)
	mux.HandleFunc("POST "+apiRoute+"chirps", apiCfg.postChirpHandler)
	mux.HandleFunc("GET "+apiRoute+"chirps", apiCfg.getChirpsHandler)

	// Admin routes
	mux.Handle("POST "+adminRoute+"reset", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiCfg.platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		apiCfg.fileserverHits.Store(0)
		if err := apiCfg.db.DeleteAllUsers(r.Context()); err != nil {
			log.Printf("Failed to delete users: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))

	mux.Handle("GET "+adminRoute+"metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, apiCfg.fileserverHits.Load())
	}))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
