//go:build !tsqlite

package main

import (
	"fmt"
	"os"
	"path"

	storage "github.com/thek4n/t/internal/storage"
)

const T_BASE_DIR = ".t"

func initTaskStorage() storage.TasksStorage {
	tBasePath, err := getBaseDir()
	if err != nil {
		die("%s", err.Error())
	}

	return &storage.FSTasksStorage{TBaseDir: tBasePath}
}

func createNamespace(namespace string) error {
	tBasePath, err := getBaseDir()
	if err != nil {
		return err
	}

	namespacePath := path.Join(tBasePath, namespace)

	return createDirectoryIfNotExists(namespacePath)
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

func cleanupEmptyNamespaces(s storage.TasksStorage) error {
	namespaces, err := s.GetNamespaces()
	if err != nil {
		return err
	}

	tBasePath, err := getBaseDir()
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		err = removeEmptyDir(path.Join(tBasePath, ns))
		if err != nil {
			return err
		}
	}

	return nil
}

func removeEmptyDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("'%s' not a directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return os.Remove(path)
	}

	return nil
}

func getBaseDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable is invalid")
	}

	return path.Join(home, T_BASE_DIR), nil
}
