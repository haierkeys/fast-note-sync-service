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

	rows, err := db.Query("SELECT id, type, password_mode, password_value FROM backup_config")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("ID | Type | PassMode | PassValue")
	fmt.Println("-------------------------------")
	for rows.Next() {
		var id int64
		var t string
		var mode int
		var val string
		rows.Scan(&id, &t, &mode, &val)
		fmt.Printf("%d | %s | %d | %s\n", id, t, mode, val)
	}
}
