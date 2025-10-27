package handlers

type UpsertRequest struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpsertResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
}

type QueryRequest struct {
	Vector []float32 `json:"vector"`
	TopK   int       `json:"top_k"`
}

type QueryResponse struct {
	Results []QueriedVector `json:"results"`
}

type QueriedVector struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type GetAllResponse struct {
	Vectors []StoredVector `json:"vectors"`
}

type StoredVector struct {
	ID             string                 `json:"id"`
	Vector         []float32              `json:"vector"`
	OriginalVector []float32              `json:"original_vector"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
