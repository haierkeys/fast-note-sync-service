package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	files, _ := filepath.Glob("storage/database/db_user_backup_*.sqlite3")
	for _, file := range files {
		fmt.Printf("Checking file: %s\n", file)
		db, err := sql.Open("sqlite3", file)
		if err != nil {
			log.Printf("Error opening %s: %v", file, err)
			continue
		}
		
		rows, err := db.Query("PRAGMA table_info(backup_config)")
		if err != nil {
			log.Printf("Error querying table_info for %s: %v", file, err)
			db.Close()
			continue
		}
		
		foundMode := false
		foundValue := false
		for rows.Next() {
			var cid int
			var name string
			var dtype string
			var notnull int
			var dfltValue interface{}
			var pk int
			rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk)
			if name == "password_mode" {
				foundMode = true
			}
			if name == "password_value" {
				foundValue = true
			}
		}
		rows.Close()
		fmt.Printf("  password_mode: %v\n", foundMode)
		fmt.Printf("  password_value: %v\n", foundValue)

		dataRows, err := db.Query("SELECT id, type, password_mode, password_value FROM backup_config")
		if err == nil {
			for dataRows.Next() {
				var id int64
				var t string
				var mode int
				var val string
				dataRows.Scan(&id, &t, &mode, &val)
				fmt.Printf("  Data: ID=%d, Type=%s, Mode=%d, Val=%s\n", id, t, mode, val)
			}
			dataRows.Close()
		} else {
			fmt.Printf("  Error reading data: %v\n", err)
		}
		db.Close()
	}
}
