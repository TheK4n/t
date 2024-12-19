package storage


import (
	"os"
	"path"
	"sort"
	"fmt"
	"strconv"
	"strings"
)


const PATH_SEPARATOR_REPLACER = "%2F"


type FSTasksStorage struct {
	tBaseDir string
}


func (ts *FSTasksStorage) GetSorted(namespace string) ([]string, error) {
	namespacePath := path.Join(ts.tBaseDir, namespace)
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
	taskContent, err := os.ReadFile(path.Join(ts.tBaseDir, namespace, taskNameToRead))
	if err != nil {
		return nil, fmt.Errorf("Error reading task: %w", err)
	}

	return taskContent, nil
}

func (ts *FSTasksStorage) GetContentByName(namespace string, name string) ([]byte, error) {
	content, err := os.ReadFile(path.Join(ts.tBaseDir, namespace, name))
	if err != nil {
		return nil, fmt.Errorf("Error reading task: %w", err)
	}

	return content, nil
}

func (ts *FSTasksStorage) DeleteTasksByIndexes(namespace string, indexes []string) error {
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
		deleteErr := os.Remove(path.Join(ts.tBaseDir, namespace, taskNameToDelete))
		if deleteErr != nil {
			return fmt.Errorf("Error remove task: %s", deleteErr)
		}
	}

	return nil
}

func (ts *FSTasksStorage) EditByName(namespace string, name string, data []byte) error {
	taskToEdit := path.Join(ts.tBaseDir, namespace, name)
	return os.WriteFile(taskToEdit, data, 0644)
}

func (ts *FSTasksStorage) Add(namespace string, name string) error {
	name = strings.ReplaceAll(name, "/", PATH_SEPARATOR_REPLACER)

	err := os.WriteFile(path.Join(ts.tBaseDir, namespace, name), []byte{}, 0644)
	if err != nil {
		return fmt.Errorf("Error write task: %s", err)
	}

	return nil
}

func (ts *FSTasksStorage) EditByIndex(namespace string, index string, data []byte) error {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return fmt.Errorf("Wrong task index")
	}

	taskIndexToEdit := tasks[taskIndex-1]
	taskToEdit := path.Join(ts.tBaseDir, namespace, taskIndexToEdit)
	return os.WriteFile(taskToEdit, data, 0644)
}