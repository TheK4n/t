//go:build tsqlite

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	storage "github.com/thek4n/t/internal/storage"
	"os"
	"path"
)

const T_BASE_DIR = ".t"

func initTaskStorage() storage.TasksStorage {
	tBasePath, err := getBaseDir()
	if err != nil {
		die("%s", err.Error())
	}

	dbPath := path.Join(tBasePath, "t.sqlite3")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		die("%s", err.Error())
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS
	tasks(
		name VARCHAR(90) NOT NULL,
		namespace VARCHAR(30) NOT NULL,
		updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		content TEXT NOT NULL,
		UNIQUE (name, namespace));
	`)

	if err != nil {
		die("%s", err.Error())
	}

	return &storage.SqlTasksStorage{DbPath: dbPath}
}

func createNamespace(_ string) error {
	return nil
}

func cleanupEmptyNamespaces(_ storage.TasksStorage) error {
	return nil
}

func getBaseDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable is invalid")
	}

	return path.Join(home, T_BASE_DIR), nil
}
