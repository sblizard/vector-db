package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sblizard/vector-db/internal/config"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server starting on port %d...\n", cfg.Server.Port)
	fmt.Printf("Using %d subspaces, vector dim: %d\n", cfg.NumSubspaces, cfg.Dim)

	http.HandleFunc("/", hello)
	http.HandleFunc("/headers", headers)

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
}
