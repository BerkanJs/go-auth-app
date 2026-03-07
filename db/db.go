package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "people.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS people (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        surname TEXT,
        email TEXT NOT NULL UNIQUE,
        age INTEGER,
        phone TEXT,
        password_hash TEXT NOT NULL
    );`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	refreshTokenTable := `CREATE TABLE IF NOT EXISTS refresh_tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        token TEXT NOT NULL UNIQUE,
        revoked INTEGER NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL,
        revoked_at TEXT
    );`

	_, err = DB.Exec(refreshTokenTable)
	if err != nil {
		log.Fatal(err)
	}

}
