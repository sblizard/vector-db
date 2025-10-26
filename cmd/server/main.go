package main

import (
	"fmt"
	"log"

	"github.com/sblizard/vector-db/internal/config"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server starting on port %d...\n", cfg.Server.Port)
	fmt.Printf("Using %d subspaces, vector dim: %d\n", cfg.NumSubspaces, cfg.Dim)
}
