//go:generate go run version_gen.go

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	handlers "github.com/thek4n/t/internal/handlers"
	"github.com/thek4n/t/internal/storage"
)

const DEFAULT_NAMESPACE = "def"
const ENVFILE = ".tns"

func main() {
	s := initTaskStorage()

	if len(os.Args) < 2 {
		namespace := getNamespace()
		err := createNamespace(namespace)
		if err != nil {
			die("Error creating namespace: %s", err)
		}

		err = handlers.ShowTasks(namespace, s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = cleanupEmptyNamespaces(s)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
		return
	}

	commands := map[string]func(storage.TasksStorage) error{
		"show": cmdShow,

		"add": cmdAdd,
		"a":   cmdAdd,

		"d":      cmdDone,
		"done":   cmdDone,
		"delete": cmdDone,

		"e":    cmdEdit,
		"edit": cmdEdit,

		"get": cmdGet,

		"ns":         cmdNamespaces,
		"namespaces": cmdNamespaces,

		"all": cmdAll,

		"-h":     cmdHelp,
		"--help": cmdHelp,

		"-v":        cmdVersion,
		"--version": cmdVersion,
	}

	cmd := os.Args[1]
	handler, found := commands[cmd]

	if !found {
		namespace := getNamespace()
		err := createNamespace(namespace)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error creating namespace: %s", err)
		}

		err = handlers.ShowTaskContentByIndex(namespace, cmd, s)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error: %s", err)
		}

		cleanupEmptyNamespaces(s)
		os.Exit(0)
	}

	err := handler(s)
	if err != nil {
		die("%s", err)
	}

	cleanupEmptyNamespaces(s)
	os.Exit(0)
}

func showVersion() error {
	_, err := fmt.Print(version)
	return err
}

func cmdShow(s storage.TasksStorage) error {
	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	return handlers.ShowTasks(namespace, s)
}

func cmdAdd(s storage.TasksStorage) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("%s", "Not enough args")
	}

	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	err = handlers.AddTask(namespace, strings.Join(os.Args[2:], " "), s)
	if err != nil {
		return fmt.Errorf("Error adding task: %s", err)
	}

	return nil
}

func cmdDone(s storage.TasksStorage) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("%s", "Not enough args")
	}

	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	err = handlers.DeleteTasksByIndexes(namespace, os.Args[2:], s)
	if err != nil {
		return fmt.Errorf("Error deleting task: %s", err)
	}

	return nil
}

func cmdEdit(s storage.TasksStorage) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("%s", "Not enough args")
	}

	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	err = handlers.EditTaskByIndex(namespace, os.Args[2], s)
	if err != nil {
		return fmt.Errorf("Error editing task: %s", err)
	}

	return nil
}

func cmdGet(s storage.TasksStorage) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("%s", "Not enough args")
	}

	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	err = handlers.ShowTaskContentByName(namespace, os.Args[2], s)
	if err != nil {
		return fmt.Errorf("Error reading task: %s", err)
	}

	return nil
}

func cmdNamespaces(s storage.TasksStorage) error {
	err := handlers.ShowNamespaces(s)
	if err != nil {
		return fmt.Errorf("Error reading namespace: %s", err)
	}

	return nil
}

func cmdAll(s storage.TasksStorage) error {
	return handlers.ShowAllTasksFromAllNamespaces(s)
}

func cmdHelp(_ storage.TasksStorage) error {
	return handlers.ShowHelp()
}

func cmdVersion(_ storage.TasksStorage) error {
	return showVersion()
}

func getNamespace() string {
	namespace, err := getNamespaceFromEnvOrFromFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s, using default namespace (%s)\n", err, DEFAULT_NAMESPACE)
		return DEFAULT_NAMESPACE
	}

	return namespace
}

func getNamespaceFromEnvOrFromFile() (string, error) {
	tEnv := os.Getenv("t")
	if tEnv != "" {
		return tEnv, nil
	}

	curdir, _ := os.Getwd()
	foundEnvFile := findFileUpTree(curdir, ENVFILE)

	if foundEnvFile == "" {
		return DEFAULT_NAMESPACE, nil
	}

	envFileContent, err := os.ReadFile(foundEnvFile)
	if err != nil {
		return "", fmt.Errorf("error reading env file: %s", foundEnvFile)
	}

	return strings.Trim(string(envFileContent), " \n"), nil
}

func findFileUpTree(startdir string, filename string) string {
	if startdir == "/" {
		return ""
	}
	if _, err := os.Stat(path.Join(startdir, filename)); err == nil {
		return path.Join(startdir, filename)
	}
	return findFileUpTree(filepath.Dir(startdir), filename)
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
