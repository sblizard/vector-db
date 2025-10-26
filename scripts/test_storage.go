package main

import (
	"fmt"

	"github.com/sblizard/vector-db/internal/storage"
)

func main() {
	layout := storage.NewLayout("./data")

	vec := []float32{0.1, 0.2, 0.3}
	_ = storage.AppendVector(layout.VectorFile(0), vec)
	fmt.Println("Vector appended")

	loaded, _ := storage.LoadVectors(layout.VectorFile(0), 3)
	fmt.Println("Loaded:", loaded)

	meta, _ := storage.OpenMetaStore(layout.MetaDB())
	defer func() {
		_ = meta.Close()
	}()

	_ = meta.Put("vector_0", []byte("cluster_0"))
	val, _ := meta.Get("vector_0")
	fmt.Println("Metadata:", string(val))
}
