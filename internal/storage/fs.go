//go:build !tsqlite

package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
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
		isNotHiddenDir := de.IsDir() && de.Name()[0] != '.'
		if isNotHiddenDir {
			result = append(result, de.Name())
		}
	}
	return result, nil
}

func (ts *FSTasksStorage) Count(namespace string) (uint, error) {
	namespaceDirEntries, err := os.ReadDir(path.Join(ts.TBaseDir, namespace))
	if err != nil {
		return 0, err
	}
	return uint(len(namespaceDirEntries)), nil
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

func (ts *FSTasksStorage) GetContentByIndex(namespace string, index string) ([]byte, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return nil, err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return nil, fmt.Errorf("Wrong task index: %s", index)
	}

	taskNameToRead := tasks[taskIndex-1]
	taskContent, err := os.ReadFile(path.Join(ts.TBaseDir, namespace, taskNameToRead))
	if err != nil {
		return nil, fmt.Errorf("Error reading task: %w", err)
	}

	return taskContent, nil
}

func (ts *FSTasksStorage) GetContentByName(namespace string, name string) ([]byte, error) {
	content, err := os.ReadFile(path.Join(ts.TBaseDir, namespace, name))
	if err != nil {
		return nil, fmt.Errorf("Error reading task: %w", err)
	}

	return content, nil
}

func (ts *FSTasksStorage) DeleteByIndexes(namespace string, indexes []string) error {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return err
	}

	for _, inputedTaskIndex := range indexes {
		taskIndex, err := strconv.Atoi(inputedTaskIndex)
		if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
			return fmt.Errorf("Wrong task index: %s", inputedTaskIndex)
		}

		taskNameToDelete := tasks[taskIndex-1]
		deleteErr := os.Remove(path.Join(ts.TBaseDir, namespace, taskNameToDelete))
		if deleteErr != nil {
			return fmt.Errorf("Error remove task: %s", deleteErr)
		}
	}

	return nil
}

func (ts *FSTasksStorage) Add(namespace string, name string) error {
	name = strings.ReplaceAll(name, "/", PATH_SEPARATOR_REPLACER)

	err := os.WriteFile(path.Join(ts.TBaseDir, namespace, name), []byte{}, 0644)
	if err != nil {
		return fmt.Errorf("Error write task: %s", err)
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

func (ts *FSTasksStorage) WriteByIndex(namespace string, index string, r io.Reader) error {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return fmt.Errorf("Wrong task index")
	}

	taskNameToEdit := tasks[taskIndex-1]
	taskToEdit := path.Join(ts.TBaseDir, namespace, taskNameToEdit)

	fileTask, err := os.Create(taskToEdit)
	if err != nil {
		return err
	}
	defer fileTask.Close()

	_, err = io.Copy(fileTask, r)
	return err
}

func (ts *FSTasksStorage) GetNameByIndex(namespace string, index string) (string, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return "", err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return "", fmt.Errorf("Wrong task index")
	}

	return tasks[taskIndex-1], nil
}

func (ts *FSTasksStorage) CountLines(namespace string, name string) (uint, error) {
	return countFileLines(path.Join(ts.TBaseDir, namespace, name))
}

func countFileLines(filePath string) (uint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1024)
	var count uint = 0
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		count += uint(bytes.Count(buf[:c], lineSep))

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
