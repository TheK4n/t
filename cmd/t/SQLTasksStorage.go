//go:build tsqlite

package main

import (
	"os"
	"path"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	storage "github.com/thek4n/t/internal/storage"
)

func initTaskStorage(namespace string) storage.TasksStorage {
	home := os.Getenv("HOME")
	if home == "" {
		die("HOME environment variable is invalid")
	}

	dbPath := path.Join(home, ".t", "t.sqlite3")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		die("%s", err.Error())
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS
	tasks(
		name varchar(90) primary key,
		namespace varchar(30) not null,
		born timestamp default CURRENT_TIMESTAMP not null,
		content text not null);
	`)

	return &storage.SqlTasksStorage{DbPath: dbPath}
}
