package main

import (
	"fmt"
	"os"
	"strings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := "storage/database/db_full.sqlite3"
	os.Remove(dbPath)
	
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect database: %v\n", err)
		return
	}
	
	sqlContent, err := os.ReadFile("scripts/db.sql")
	if err != nil {
		fmt.Printf("Failed to read scripts/db.sql: %v\n", err)
		return
	}
	
	// Split by semicolon and execute each statement
	// This is a bit naive but should work for the provided db.sql
	queries := strings.Split(string(sqlContent), ";")
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" || strings.HasPrefix(q, "--") {
			continue
		}
		if err := db.Exec(q).Error; err != nil {
			fmt.Printf("Error executing: %s\nError: %v\n", q, err)
		}
	}
	
	fmt.Println("Full database created successfully at storage/database/db_full.sqlite3")
}
