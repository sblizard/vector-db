package storage

import (
	"fmt"
	"path/filepath"
)

type Layout struct {
	BasePath string
}

func NewLayout(basePath string) *Layout {
	return &Layout{BasePath: basePath}
}

func (l *Layout) VectorFile(subspaceID int) string {
	return filepath.Join(l.BasePath, fmt.Sprintf("cluster_%d.bin", subspaceID))
}

func (l *Layout) CentroidFile() string {
	return filepath.Join(l.BasePath, "centroids.bin")
}

func (l *Layout) MetaDB() string {
	return filepath.Join(l.BasePath, "metadata.db")
}
