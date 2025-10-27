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

	handler := api.NewHandler(cfg.Engine)
	upsertHandler := handlers.NewUpsertHandler(cfg.Engine)
	readHandler := handlers.NewReadHandler(cfg.Engine)
	deleteHandler := handlers.NewDeleteHandler(cfg.Engine)
	searchHandler := handlers.NewSearchHandler(cfg.Engine)

	router := handlers.NewRouter(handler, upsertHandler, readHandler, deleteHandler, searchHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Server running on %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, router))
}
