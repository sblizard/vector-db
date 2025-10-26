package api

import (
	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", handler.Health).Methods("GET")
	r.HandleFunc("/upsert", handler.Upsert).Methods("POST")
	r.HandleFunc("/hello", handler.Hello).Methods("GET")
	r.HandleFunc("/headers", handler.Headers).Methods("GET")
	r.HandleFunc("/vectors", handler.GetAllVectors).Methods("GET")

	return r
}
