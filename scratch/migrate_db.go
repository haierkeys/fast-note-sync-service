package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "storage/database/db_full.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := []string{
		"ALTER TABLE backup_config ADD COLUMN password_mode INTEGER DEFAULT 0;",
		"ALTER TABLE backup_config ADD COLUMN password_value TEXT DEFAULT '';",
	}

	for _, q := range queries {
		fmt.Printf("Executing: %s\n", q)
		_, err := db.Exec(q)
		if err != nil {
			fmt.Printf("Error (might be okay if column exists): %v\n", err)
		} else {
			fmt.Println("Success")
		}
	}
}
