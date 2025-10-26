package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received upsert: ID=%s, VectorDim=%d\n", req.ID, len(req.Vector))

	resp := UpsertResponse{Status: "success", Message: "vector received"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func (h *Handler) Headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
