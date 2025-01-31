//go:generate go run version_gen.go

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	storage "github.com/thek4n/t/internal/storage"
)

const DEFAULT_NAMESPACE = "def"
const PATH_SEPARATOR_REPLACER = "%2F"
const ENVFILE = ".tns"
const HELP_MESSAGE = `T simple task tracker

USAGE
    t                            - Show tasks in format '[INDEX] TASK NAME (LINES)'
    t get (TASK)                 - Get task content
    t show                       - Show tasks in format '[INDEX] TASK NAME (LINES)'
    t (INDEX)                    - Show task content
    t add (X X X)                - Add task with name X X X
    t edit (INDEX)               - Edit task with INDEX by \$EDITOR
    t done (INDEX) [INDEX] ...   - Delete tasks with INDEXes
    t namespaces                 - Show namespaces
    t --help                     - Show this message
    t --version                  - Show version

    t a       - alias for add
    t e       - alias for edit
    t d       - alias for done
    t delete  - alias for done
    t ns      - alias for namespaces

NAMESPACES
    t namespaces             # show namespaces
    t=work t a fix bug 211   # add task in workspace 'work'
    t=work t                 # show tasks in workspace 'work'

NAMESPACE FILE
    File with name '.tns' can be in current directory or any directory up the tree
    File contains name of namespace
    Environment variable 't' overwrite using this file

    Example:
    $ cat .tns
    dotfiles
    $ t show
    # dotfiles
    ...
    $ t=storage t
    # storage
    ...
`

func main() {
	s := initTaskStorage()

	if len(os.Args) < 2 {
		namespace := getNamespace()
		err := createNamespace(namespace)
		if err != nil {
			die("Error creating namespace: %s", err)
		}

		err = showTasks(namespace, s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = cleanupEmptyNamespaces(s)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}

	handlers := map[string]func(storage.TasksStorage) error{
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
	handler, found := handlers[cmd]

	if !found {
		namespace := getNamespace()
		err := createNamespace(namespace)
		if err != nil {
			cleanupEmptyNamespaces(s)
			die("Error creating namespace: %s", err)
		}

		err = showTaskContentByIndex(namespace, cmd, s)
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

func cmdShow(s storage.TasksStorage) error {
	namespace := getNamespace()
	err := createNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error creating namespace: %s", err)
	}

	return showTasks(namespace, s)
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

	err = addTask(namespace, strings.Join(os.Args[2:], " "), s)
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

	err = deleteTasksByIndexes(namespace, os.Args[2:], s)
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

	err = editTaskByIndex(namespace, os.Args[2], s)
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

	err = showTaskContentByName(namespace, os.Args[2], s)
	if err != nil {
		return fmt.Errorf("Error reading task: %s", err)
	}

	return nil
}

func cmdNamespaces(s storage.TasksStorage) error {
	err := showNamespaces(s)
	if err != nil {
		return fmt.Errorf("Error reading namespace: %s", err)
	}

	return nil
}

func cmdAll(s storage.TasksStorage) error {
	return ShowAllTasksFromAllNamespaces(s)
}

func cmdHelp(_ storage.TasksStorage) error {
	return showHelp()
}

func cmdVersion(_ storage.TasksStorage) error {
	return showVersion()
}

func showTasks(namespace string, s storage.TasksStorage) error {
	tasks, err := s.GetSorted(namespace)
	if err != nil {
		return err
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n", namespace)
	for i, task := range tasks {
		taskLines, _ := s.CountLines(namespace, task)

		formattedTaskLines := formatLinesCount(taskLines)

		formattedTaskName := strings.ReplaceAll(task, PATH_SEPARATOR_REPLACER, "/")
		fmt.Printf("[%d] %s (%s)\n", i+1, formattedTaskName, formattedTaskLines)
	}

	return nil
}

func formatLinesCount(lines uint) string {
	if lines > 70 {
		return "..."
	}
	if lines == 0 {
		return "-"
	}
	return fmt.Sprint(lines + 1)
}

func addTask(namespace string, name string, s storage.TasksStorage) error {
	return s.Add(namespace, name)
}

func deleteTasksByIndexes(namespace string, indexes []string, s storage.TasksStorage) error {
	return s.DeleteByIndexes(namespace, indexes)
}

func editTaskByIndex(namespace string, index string, s storage.TasksStorage) error {
	content, err := s.GetContentByIndex(namespace, index)
	if err != nil {
		return err
	}

	taskName, err := s.GetNameByIndex(namespace, index)
	if err != nil {
		return err
	}

	tempFile, err := createTempFile(fmt.Sprintf("t_%s_", taskName))
	if err != nil {
		return err
	}

	_, err = tempFile.Write(content) // write original text from task
	if err != nil {
		return err
	}
	tempFile.Close() // close now, because of editor

	cmd := exec.Command(os.Getenv("EDITOR"), tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error run EDITOR: %w", err)
	}

	tempFile, err = os.Open(tempFile.Name()) // reopen tempfile for reading
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	return s.WriteByName(namespace, taskName, tempFile)
}

func createTempFile(pattern string) (*os.File, error) {
	tempDir := "/tmp"
	if !exists(tempDir) {
		tempDir = os.Getenv("TMPDIR")
	}

	if tempDir == "" {
		tempDir = "."
	}

	tempFile, err := os.CreateTemp(tempDir, pattern)
	if err != nil {
		return nil, err
	}
	return tempFile, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		return false
	}
	return true
}

func showTaskContentByName(namespace string, name string, s storage.TasksStorage) error {
	content, err := s.GetContentByName(namespace, name)
	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

func showNamespaces(s storage.TasksStorage) error {
	nss, err := s.GetNamespaces()

	if err != nil {
		return err
	}

	for _, ns := range nss {
		namespaceTasksCount, err := s.Count(ns)
		if err != nil {
			fmt.Printf("%s (%s)\n", ns, "-")
			continue
		}
		fmt.Printf("%s (%d)\n", ns, namespaceTasksCount)
	}
	return nil
}

func showHelp() error {
	_, err := fmt.Print(HELP_MESSAGE)
	return err
}

func showVersion() error {
	_, err := fmt.Print(version)
	return err
}

func showTaskContentByIndex(namespace string, index string, s storage.TasksStorage) error {
	taskContent, err := s.GetContentByIndex(namespace, index)
	taskName, err := s.GetNameByIndex(namespace, index)

	if err != nil {
		return err
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n\n", taskName)
	fmt.Print(string(taskContent))
	return nil
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

func ShowAllTasksFromAllNamespaces(s storage.TasksStorage) error {
	namespaces, err := s.GetNamespaces()
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		currentNamespaceTasks, err := s.GetSorted(namespace)
		if err != nil {
			return err
		}
		for _, task := range currentNamespaceTasks {
			fmt.Printf("[%s] %s\n", namespace, task)
		}
	}
	return nil
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
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
