package models

import (
	"encoding/json"
	"os"
)

type RetryLevel struct {
	QuerySuffix string `json:"query_suffix"`
	Tolerance   int    `json:"tolerance"`
}

type Config struct {
	DownloadPath string       `json:"download_path"`
	Workers      int          `json:"workers"`
	Retries      []RetryLevel `json:"retries"`
	Debug        bool         `json:"debug_mode"`
	LightMode    bool         `json:"light_mode"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	//Convert struct to JSON (with 4-space indent for readability)
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	//Write the bytes to the file (Permissions: 0644)
	return os.WriteFile(path, data, 0644)
}
