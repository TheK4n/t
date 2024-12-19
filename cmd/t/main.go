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

const T_BASE_DIR = ".t"
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
	home := os.Getenv("HOME")
	if home == "" {
		die("HOME environment variable is invalid")
	}

	namespace, err := getNamespaceFromEnvOrFromFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s, using default namespace (%s)\n", err, DEFAULT_NAMESPACE)
	}

	namespacePath := path.Join(home, T_BASE_DIR, namespace)

	createNamespaceErr := createDirectoryIfNotExists(namespacePath)
	if createNamespaceErr != nil {
		die("Error creating namespace: %s", createNamespaceErr)
	}

	tBasePath := path.Join(home, T_BASE_DIR)

	s := &storage.FSTasksStorage{TBaseDir: tBasePath}

	if len(os.Args) < 2 {
		showTasks(namespace, s)
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "show":
		showTasks(namespace, s)
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "a", "add":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := addTask(namespace, strings.Join(os.Args[2:], " "), s)
		if err != nil {
			die("Error adding task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "d", "done", "delete":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := deleteTasksByIndexes(namespace, os.Args[2:], s)
		if err != nil {
			die("Error deleting task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "e", "edit":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := editTaskByIndex(namespace, os.Args[2], s)
		if err != nil {
			die("Error editing task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "get":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := showTaskContentByName(namespace, os.Args[2], s)
		if err != nil {
			die("Error reading task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "ns", "namespaces":
		err := showNamespaces(s)
		if err != nil {
			die("Error reading namespace: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "-h", "--help":
		showHelp()

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "-v", "--version":
		showVersion()

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	default:
		err := showTaskContentByIndex(namespace, cmd, s)
		if err != nil {
			die("Error: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}
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
		return DEFAULT_NAMESPACE, fmt.Errorf("error reading env file: %s", foundEnvFile)
	}

	return strings.Trim(string(envFileContent), " \n"), nil
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

	tempFile, err := os.CreateTemp("/tmp", fmt.Sprintf("t_%s_", taskName))
	if err != nil {
		return err
	}

	_, err = tempFile.Write(content)
	if err != nil {
		return err
	}
	tempFile.Close()

	cmd := exec.Command(os.Getenv("EDITOR"), tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error run EDITOR: %w", err)
	}

	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	return s.WriteByIndex(namespace, index, tempFile)
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

func showHelp() {
	fmt.Print(HELP_MESSAGE)
}

func showVersion() {
	fmt.Print(version)
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

func removeEmptyNamespaces(dir string) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		die("Error reading namespace to remove: %s", err)
	}

	for _, de := range dirEntries {
		subdirEntries, err := os.ReadDir(path.Join(dir, de.Name()))
		if err != nil {
			continue
		}
		if len(subdirEntries) < 1 {
			rmErr := os.Remove(path.Join(dir, de.Name()))
			if rmErr != nil {
				die("Error remove namespace: %s", rmErr)
			}
		}
	}
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
