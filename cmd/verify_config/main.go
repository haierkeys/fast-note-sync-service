package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
)

func main() {
	configPath := "config/config.yaml"
	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("Loading config from: %s\n", absPath)

	cfg, _, err := app.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("Storage Configuration Loaded:")
	fmt.Printf("LocalFS Enabled: %v\n", cfg.Storage.LocalFS.IsEnabled)
	fmt.Printf("LocalFS HttpfsIsEnable: %v\n", cfg.Storage.LocalFS.HttpfsIsEnable)
	fmt.Printf("LocalFS SavePath: %s\n", cfg.Storage.LocalFS.SavePath)

	fmt.Printf("AliyunOSS Enabled: %v\n", cfg.Storage.AliyunOSS.IsEnabled)
	fmt.Printf("AwsS3 Enabled: %v\n", cfg.Storage.AwsS3.IsEnabled)

	if !cfg.Storage.LocalFS.IsEnabled {
		log.Fatal("LocalFS should be enabled")
	}
	if cfg.Storage.LocalFS.SavePath != "storage/uploads" {
		log.Fatalf("LocalFS SavePath mismatch, expected 'storage/uploads', got '%s'", cfg.Storage.LocalFS.SavePath)
	}
}
