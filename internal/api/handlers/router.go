package handlers

import (
	"github.com/gorilla/mux"
	"github.com/sblizard/vector-db/internal/api"
)

func NewRouter(healthHandler *api.Handler, upsertHandler *UpsertHandler, readHandler *ReadHandler) *mux.Router {
	r := mux.NewRouter()

	// Health check routes
	r.HandleFunc("/health", healthHandler.Health).Methods("GET")
	r.HandleFunc("/hello", healthHandler.Hello).Methods("GET")
	r.HandleFunc("/headers", healthHandler.Headers).Methods("GET")

	// Vector operation routes
	r.HandleFunc("/upsert", upsertHandler.Upsert).Methods("POST")
	r.HandleFunc("/vectors", readHandler.GetAllVectors).Methods("GET")

	return r
}
