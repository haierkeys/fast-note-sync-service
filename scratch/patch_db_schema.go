package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "storage/database/db_full.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sql := `
    DROP TABLE IF EXISTS "auth_token_log";
    CREATE TABLE "auth_token_log" (
        "id" integer PRIMARY KEY AUTOINCREMENT,
        "token_id" integer NOT NULL DEFAULT 0,
        "uid" integer NOT NULL DEFAULT 0,
        "protocol" text NOT NULL DEFAULT '',
        "client_type" text NOT NULL DEFAULT '',
        "client_name" text NOT NULL DEFAULT '',
        "client_version" text NOT NULL DEFAULT '',
        "path" text NOT NULL DEFAULT '',
        "method" text NOT NULL DEFAULT '',
        "ip" text NOT NULL DEFAULT '',
        "status_code" integer NOT NULL DEFAULT 0,
        "created_at" datetime DEFAULT NULL
    );`

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database schema updated successfully.")
}
