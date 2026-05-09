package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("storage/database/db.sqlite3"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect database: %v\n", err)
		return
	}
	
	err = db.Exec("ALTER TABLE auth_token ADD COLUMN issue_type INTEGER NOT NULL DEFAULT 1").Error
	if err != nil {
		fmt.Printf("Failed to add column: %v\n", err)
	} else {
		fmt.Println("Column issue_type added successfully to auth_token table.")
	}
}
