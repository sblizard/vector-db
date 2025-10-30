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

	_, _, engine, err := cfg.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	handler := api.NewHandler(engine)
	upsertHandler := handlers.NewUpsertHandler(engine)
	readHandler := handlers.NewReadHandler(engine)
	deleteHandler := handlers.NewDeleteHandler(engine)
	searchHandler := handlers.NewSearchHandler(engine)

	router := handlers.NewRouter(handler, upsertHandler, readHandler, deleteHandler, searchHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Server running on %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, router))
}
