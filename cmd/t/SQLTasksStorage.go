//go:build tsqlite

package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	storage "github.com/thek4n/t/internal/storage"
)

const DB_PATH = "./test.sqlite3"

func initTaskStorage(namespace string) storage.TasksStorage {
	db, err := sql.Open("sqlite3", DB_PATH)
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

	return &storage.SqlTasksStorage{DbPath: DB_PATH}
}
