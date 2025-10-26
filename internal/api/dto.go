package api

type UpsertRequest struct {
	ID       string            `json:"id"`
	Vector   []float32         `json:"vector"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type UpsertResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type QueryRequest struct {
	Vector []float32 `json:"vector"`
	TopK   int       `json:"top_k"`
}

type QueryResponse struct {
	Results []QueryResult `json:"results"`
}

type QueryResult struct {
	ID       string            `json:"id"`
	Score    float32           `json:"score"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
