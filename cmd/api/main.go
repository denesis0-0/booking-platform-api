package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/denesis0-0/booking-platform-api/internal/config"
	"github.com/denesis0-0/booking-platform-api/internal/storage"
)

type HealthResponse struct {
	Status   string `json:"status"`
	App      string `json:"app"`
	Database string `json:"database"`
}

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := storage.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		healthHandler(w, r, db)
	})

	addr := ":" + cfg.Port

	log.Printf("server started on http://localhost%s", addr)

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request, db *storage.Postgres) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	databaseStatus := "ok"
	statusCode := http.StatusOK

	if err := db.Ping(ctx); err != nil {
		databaseStatus = "error"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:   "ok",
		App:      "booking-platform-api",
		Database: databaseStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
