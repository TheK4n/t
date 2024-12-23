//go:build !tsqlite
package main

import (
	"os"
	"path"

	storage "github.com/thek4n/t/internal/storage"
)

func getTaskStorage() storage.TasksStorage {
	home := os.Getenv("HOME")
	if home == "" {
		die("HOME environment variable is invalid")
	}

	tBasePath := path.Join(home, T_BASE_DIR)

	return &storage.FSTasksStorage{TBaseDir: tBasePath}
}
