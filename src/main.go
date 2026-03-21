package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/SirLouen/chirpy-bootdev/src/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const port = "8080"
	const filepathRoot = "./"
	const apiRoute = "/api/"
	const adminRoute = "/admin/"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer db.Close()
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:             dbQueries,
		platform:       platform,
		fileserverHits: atomic.Int32{},
		secret:         secret,
	}
	mux := http.NewServeMux()

	// Static file server for /app/ route
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))

	// API Routes
	mux.HandleFunc("GET "+apiRoute+"healthz", handlerHealthz)

	mux.HandleFunc("POST "+apiRoute+"users", apiCfg.handlerUsersCreate)

	mux.HandleFunc("POST "+apiRoute+"chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET "+apiRoute+"chirps", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET "+apiRoute+"chirps/{chirpID}", apiCfg.handlerChirpGet)

	mux.HandleFunc("POST "+apiRoute+"login", apiCfg.loginHandler)

	// Admin routes
	mux.HandleFunc("POST "+adminRoute+"reset", apiCfg.handlerReset)
	mux.HandleFunc("GET "+adminRoute+"metrics", apiCfg.handlerMetrics)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
