package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sblizard/vector-db/internal/api"
	"github.com/sblizard/vector-db/internal/api/handlers"
	"github.com/sblizard/vector-db/internal/config"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handler := api.NewHandler(cfg.Storage, cfg.Layout)
	upsertHandler := handlers.NewUpsertHandler(cfg.Storage, cfg.Layout)
	readHandler := handlers.NewReadHandler(cfg.Storage, cfg.Layout)

	router := handlers.NewRouter(handler, upsertHandler, readHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Server running on %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, router))
}
