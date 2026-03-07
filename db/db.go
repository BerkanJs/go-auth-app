package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite", "people.db")
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
        photo_path TEXT,
        role TEXT NOT NULL DEFAULT 'editor',
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

	blogTable := `CREATE TABLE IF NOT EXISTS blogs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        summary TEXT,
        image_path TEXT,
        author_id INTEGER NOT NULL,
        author_name TEXT NOT NULL,
        published INTEGER NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        FOREIGN KEY (author_id) REFERENCES people(id)
    );`

	_, err = DB.Exec(blogTable)
	if err != nil {
		log.Fatal(err)
	}

}
