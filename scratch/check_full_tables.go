package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("storage/database/db_full.sqlite3"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect database: %v\n", err)
		return
	}
	
	tables, err := db.Migrator().GetTables()
	if err != nil {
		fmt.Printf("Failed to get tables: %v\n", err)
		return
	}
	
	fmt.Printf("Tables in db_full (%d):\n", len(tables))
	for _, t := range tables {
		fmt.Println("-", t)
	}
}
