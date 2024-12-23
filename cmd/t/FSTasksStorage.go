//go:build !tsqlite

package main

import (
	"fmt"
	"os"
	"path"

	storage "github.com/thek4n/t/internal/storage"
)

const T_BASE_DIR = ".t"

func initTaskStorage(namespace string) storage.TasksStorage {
	home := os.Getenv("HOME")
	if home == "" {
		die("HOME environment variable is invalid")
	}

	namespacePath := path.Join(home, T_BASE_DIR, namespace)

	createNamespaceErr := createDirectoryIfNotExists(namespacePath)
	if createNamespaceErr != nil {
		die("Error creating namespace: %s", createNamespaceErr)
	}

	tBasePath := path.Join(home, T_BASE_DIR)

	return &storage.FSTasksStorage{TBaseDir: tBasePath}
}

func createDirectoryIfNotExists(directory string) error {
	fstat, err := os.Stat(directory)

	if err != nil {
		mkdirError := os.MkdirAll(directory, 0755)
		if mkdirError != nil {
			return fmt.Errorf("Cant create directory: %s", mkdirError)
		}
		return nil
	}

	if !fstat.IsDir() {
		return fmt.Errorf("Error: file %s already exists, and its not a directory", directory)
	}

	return nil
}
