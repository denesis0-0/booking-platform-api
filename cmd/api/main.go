package main

import (
	"context"
	"log"
	"net/http"

	"github.com/denesis0-0/booking-platform-api/internal/config"
	"github.com/denesis0-0/booking-platform-api/internal/httpapi"
	"github.com/denesis0-0/booking-platform-api/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := storage.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	handler := httpapi.NewHandler(db)
	router := handler.Routes()

	addr := ":" + cfg.Port

	log.Printf("server started on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
