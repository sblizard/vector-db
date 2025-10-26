package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sblizard/vector-db/internal/storage"
)

type ServerConfig struct {
	Port int `json:"port"`
}

type Config struct {
	Dim          int                `json:"dim"`
	NumSubspaces int                `json:"num_subspaces"`
	TopK         int                `json:"top_k"`
	TopL         int                `json:"top_l"`
	DataPath     string             `json:"data_path"`
	DBPath       string             `json:"db_path"`
	Server       ServerConfig       `json:"server"`
	Storage      *storage.MetaStore `json:"-"`
	Layout       *storage.Layout    `json:"-"`
}

func DefaultConfig() Config {
	return Config{
		Dim:          128,
		NumSubspaces: 16,
		TopK:         10,
		TopL:         2,
		DataPath:     "./data",
		DBPath:       "./data/metadata.db",
		Server:       ServerConfig{Port: 8080},
		Storage:      storage.NewMetaStore("./data/metadata.db"),
		Layout:       storage.NewLayout("./data"),
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Config file not found, using defaults: %v\n", err)
		return cfg, nil
	}
	defer func() { _ = file.Close() }()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
