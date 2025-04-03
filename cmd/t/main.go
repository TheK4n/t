//go:generate go run version_gen.go

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	handlers "github.com/thek4n/t/internal/handlers"
	"github.com/thek4n/t/internal/storage"
)

const DEFAULT_NAMESPACE = "def"
const ENVFILE = ".tns"

var COMMANDS = map[string]func(storage.TasksStorage, []string, string) error{
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

func main() {
	osArgs := os.Args[1:] // reject program name

	s := initTaskStorage()

	argsEmpty := len(osArgs) < 1
	if argsEmpty {
		namespace := getNamespace()
		err := showTasks(s, namespace)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error show namespaces: %s", err)
		}
		os.Exit(0)
	}

	firstArgumentIsWord, _ := regexp.MatchString(`[a-zA-Z]+`, osArgs[0])
	_, firstArgumentIsCommand := COMMANDS[osArgs[0]]

	firstArgumentIsNamespace := firstArgumentIsWord && !firstArgumentIsCommand

	var namespace string
	if firstArgumentIsNamespace {
		namespace = osArgs[0]
		osArgs = osArgs[1:] // reject namespace from args
	} else {
		namespace = getNamespace()
	}

	argsEmpty = len(osArgs) < 1
	if argsEmpty {
		err := showTasks(s, namespace)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error show namespaces: %s", err)
		}
		os.Exit(0)
	}

	err := createNamespace(namespace)
	if err != nil {
		die("Error creating namespace: %s", err)
	}

	commandArgumentIsNumber, _ := regexp.MatchString(`[0-9]+`, osArgs[0])
	if commandArgumentIsNumber {
		err := createNamespace(namespace)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error creating namespace: %s", err)
		}

		index, err := strconv.Atoi(osArgs[0])
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error parse index")
		}

		err = handlers.ShowTaskContentByIndex(namespace, index, s)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error: %s", err)
		}

		cleanupEmptyNamespaces(s)
		os.Exit(0)
	}

	handler, found := COMMANDS[osArgs[0]]
	if !found {
		die("Command '%s' not found", osArgs[0])
	}

	err = createNamespace(namespace)
	if err != nil {
		cleanupEmptyNamespaces(s)
		die("Error creating namespace: %s", err)
	}

	err = handler(s, osArgs[1:], namespace)
	if err != nil {
		die("Error on command '%s': %s", osArgs[0], err)
	}

	cleanupEmptyNamespaces(s)
	os.Exit(0)
}

func showTasks(s storage.TasksStorage, namespace string) error {
	err := createNamespace(namespace)
	if err != nil {
		return err
	}

	return handlers.ShowTasks(namespace, s)
}

func showVersion() error {
	_, err := fmt.Print(version)
	return err
}

func cmdShow(s storage.TasksStorage, _ []string, namespace string) error {
	return handlers.ShowTasks(namespace, s)
}

func cmdAdd(s storage.TasksStorage, args []string, namespace string) error {
	if len(args) < 1 {
		return fmt.Errorf("%s", "Not enough args")
	}

	err := handlers.AddTask(namespace, strings.Join(args, " "), s)
	if err != nil {
		return fmt.Errorf("Error adding task: %s", err)
	}

	return nil
}

func cmdDone(s storage.TasksStorage, args []string, namespace string) error {
	if len(args) < 1 {
		return fmt.Errorf("%s", "Not enough args")
	}

	indexes, err := atoiIndexes(args)
	if err != nil {
		return fmt.Errorf("Error parse indexes: %s", err)
	}

	err = handlers.DeleteTasksByIndexes(namespace, indexes, s)
	if err != nil {
		return fmt.Errorf("Error deleting task: %s", err)
	}

	return nil
}

func atoiIndexes(indexes []string) ([]int, error) {
	var res []int

	for _, index := range indexes {
		idx, err := strconv.Atoi(index)
		if err != nil {
			return nil, fmt.Errorf("Error parse index %s: %s", index, err)
		}
		res = append(res, idx)
	}

	return res, nil
}

func cmdEdit(s storage.TasksStorage, args []string, namespace string) error {
	if len(args) < 1 {
		return fmt.Errorf("%s", "Not enough args")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("Error parse index %s: %s", args[0], err)
	}

	err = handlers.EditTaskByIndex(namespace, index, s)
	if err != nil {
		return fmt.Errorf("Error editing task: %s", err)
	}

	return nil
}

func cmdGet(s storage.TasksStorage, args []string, namespace string) error {
	if len(args) < 1 {
		return fmt.Errorf("%s", "Not enough args")
	}

	err := handlers.ShowTaskContentByName(namespace, args[0], s)
	if err != nil {
		return fmt.Errorf("Error reading task: %s", err)
	}

	return nil
}

func cmdNamespaces(s storage.TasksStorage, _ []string, _ string) error {
	err := handlers.ShowNamespaces(s)
	if err != nil {
		return fmt.Errorf("Error reading namespace: %s", err)
	}

	return nil
}

func cmdAll(s storage.TasksStorage, _ []string, _ string) error {
	return handlers.ShowAllTasksFromAllNamespaces(s)
}

func cmdHelp(_ storage.TasksStorage, _ []string, _ string) error {
	return handlers.ShowHelp()
}

func cmdVersion(_ storage.TasksStorage, _ []string, _ string) error {
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
