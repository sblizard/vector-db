package api

import (
	"fmt"
	"net/http"

	"github.com/sblizard/vector-db/internal/engine"
)

type Handler struct {
	engine *engine.Engine
}

func NewHandler(engine *engine.Engine) *Handler {
	return &Handler{
		engine: engine,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"Status":"ok","Message":"Service is healthy","StatusCode":200}`))
}

func (h *Handler) Hello(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprint(w, "Hello, World!")
}

func (h *Handler) Headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
