package engine

type GetAllResponse struct {
	Vectors []StoredVector `json:"vectors"`
}

type StoredVector struct {
	ID             string                 `json:"id"`
	Vector         []float32              `json:"vector"`
	OriginalVector []float32              `json:"original_vector,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type SearchResult struct {
	ID     string    `json:"id"`
	Score  float32   `json:"score"`
	Vector []float32 `json:"vector,omitempty"`
}
