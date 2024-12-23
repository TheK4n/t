//go:build tsqlite
package main

import (
	storage "github.com/thek4n/t/internal/storage"
)


func getTaskStorage() storage.TasksStorage {
	return &storage.SqlTasksStorage{DbPath: "./test.sqlite3"}
}
