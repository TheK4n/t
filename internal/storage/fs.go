//go:build !tsqlite

package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

const PATH_SEPARATOR_REPLACER = "%2F"

type FSTasksStorage struct {
	TBaseDir string
}

func (ts *FSTasksStorage) GetNamespaces() ([]string, error) {
	dirEntries, err := os.ReadDir(ts.TBaseDir)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(dirEntries))
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		if de.Name()[0] == '.' {
			continue
		}

		result = append(result, de.Name())
	}
	return result, nil
}

func (ts *FSTasksStorage) Count(namespace string) (int, error) {
	namespaceDirEntries, err := os.ReadDir(path.Join(ts.TBaseDir, namespace))
	if err != nil {
		return 0, err
	}
	return len(namespaceDirEntries), nil
}

func (ts *FSTasksStorage) GetSorted(namespace string) ([]string, error) {
	namespacePath := path.Join(ts.TBaseDir, namespace)
	dirEntries, err := os.ReadDir(namespacePath)
	if err != nil {
		return nil, err
	}

	sortErr := sortTasks(dirEntries)
	if sortErr != nil {
		return nil, fmt.Errorf("Error sorting tasks: %s", sortErr)
	}

	result := make([]string, len(dirEntries))
	for i, de := range dirEntries {
		if de.IsDir() {
			continue
		}

		result[i] = de.Name()
	}

	return result, nil
}

func sortTasks(tasks []os.DirEntry) error {
	var sortErr error

	sort.Slice(tasks, func(i, j int) bool {
		iInfo, err := tasks[i].Info()
		jInfo, err := tasks[j].Info()

		if err != nil {
			sortErr = err
		}

		return iInfo.ModTime().Unix() > jInfo.ModTime().Unix()
	})
	return sortErr
}

func (ts *FSTasksStorage) GetContentByIndex(namespace string, index int) ([]byte, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return nil, err
	}

	if index > len(tasks) || index < 1 {
		return nil, fmt.Errorf("Wrong task index: %d", index)
	}

	taskNameToRead := tasks[index-1]
	taskContent, err := os.ReadFile(path.Join(ts.TBaseDir, namespace, taskNameToRead))
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %w", err)
	}

	return taskContent, nil
}

func (ts *FSTasksStorage) GetContentByName(namespace string, name string) ([]byte, error) {
	content, err := os.ReadFile(path.Join(ts.TBaseDir, namespace, name))
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %w", err)
	}

	return content, nil
}

func (ts *FSTasksStorage) DeleteByIndexes(namespace string, indexes []int) error {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return err
	}

	for _, inputedTaskIndex := range indexes {
		if inputedTaskIndex > len(tasks) || inputedTaskIndex < 1 {
			return fmt.Errorf("Wrong task index: %d", inputedTaskIndex)
		}

		taskNameToDelete := tasks[inputedTaskIndex-1]
		deleteErr := os.Remove(path.Join(ts.TBaseDir, namespace, taskNameToDelete))
		if deleteErr != nil {
			return fmt.Errorf("Error remove file: %s", deleteErr)
		}
	}

	return nil
}

func (ts *FSTasksStorage) Add(namespace string, name string) error {
	name = strings.ReplaceAll(name, "/", PATH_SEPARATOR_REPLACER)

	err := os.WriteFile(path.Join(ts.TBaseDir, namespace, name), []byte{}, 0644)
	if err != nil {
		return fmt.Errorf("Error write file: %s", err)
	}

	return nil
}

func (ts *FSTasksStorage) WriteByName(namespace string, name string, r io.Reader) error {
	taskToEdit := path.Join(ts.TBaseDir, namespace, name)

	file, err := os.Create(taskToEdit)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return err
}

func (ts *FSTasksStorage) WriteByIndex(namespace string, index int, r io.Reader) error {
	taskNameToEdit, err := ts.GetNameByIndex(namespace, index)
	if err != nil {
		return err
	}

	return ts.WriteByName(namespace, taskNameToEdit, r)
}

func (ts *FSTasksStorage) GetNameByIndex(namespace string, index int) (string, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return "", err
	}

	if index > len(tasks) || index < 1 {
		return "", fmt.Errorf("Wrong task index")
	}

	return tasks[index-1], nil
}

func (ts *FSTasksStorage) CountLines(namespace string, name string) (int, error) {
	return countFileLines(path.Join(ts.TBaseDir, namespace, name))
}

func countFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1024)
	var count int = 0
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
