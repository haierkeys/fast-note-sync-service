package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"gorm.io/gorm"
)

func main() {
	configPath := "config/config.yaml"
	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("Loading config from: %s\n", absPath)

	cfg, _, err := app.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Mock DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Initialize DAO and Repo
	d := dao.New(db, context.Background())
	repo := dao.NewStorageRepository(d)

	// Initialize Service
	svc := service.NewStorageService(repo, &cfg.Storage)

	// Test GetEnabledTypes
	types, err := svc.GetEnabledTypes()
	if err != nil {
		log.Fatalf("GetEnabledTypes failed: %v", err)
	}
	fmt.Printf("Enabled Types: %v\n", types)

	// Test CreateOrUpdate with enabled type
	fmt.Println("Testing CreateOrUpdate with enabled type (localfs)...")
	// Note: We need to ensure localfs is enabled in config for this test to pass validation
	// Based on previous steps, localfs might be disabled or enabled. config.yaml showed is-enable: false in one diff, but we are loading from file.
	// Let's check config first
	if cfg.Storage.LocalFS.IsEnabled {
		_, err = svc.CreateOrUpdate(context.Background(), 1, 0, &dto.StorageDTO{Type: "localfs"})
		if err != nil {
			// It might fail due to DB error (mock DB not migrated), but shouldn't be ErrorStorageTypeDisabled
			fmt.Printf("CreateOrUpdate (localfs) result: %v\n", err)
		}
	} else {
		fmt.Println("LocalFS is disabled in config, skipping positive test.")
	}

	// Test CreateOrUpdate with disabled type
	// Let's assume a type is disabled. If all are enabled, we can't test this easily without modifying config object.
	// Let's temporarily disable a type in the config object
	fmt.Println("Testing CreateOrUpdate with disabled type (mock_disabled)...")
	cfg.Storage.AliyunOSS.IsEnabled = false
	_, err = svc.CreateOrUpdate(context.Background(), 1, 0, &dto.StorageDTO{Type: "oss"})
	if err == nil {
		log.Fatal("Expected error for disabled type, got nil")
	}
	fmt.Printf("CreateOrUpdate (oss-disabled) result: %v\n", err)

}
