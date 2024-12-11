//go:generate go run version_gen.go

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
    t=work t                 # show tasks in workspace 'work'`

func main() {
	home := os.Getenv("HOME")

	var ns string

	if os.Getenv("t") == "" {
		curdir, _ := os.Getwd()

		foundEnvFile := findFileUpTree(curdir, ENVFILE)
		if foundEnvFile != "" {
			if _, err := os.Stat(foundEnvFile); err == nil {
				envFileContent, err := os.ReadFile(foundEnvFile)
				if err == nil {
					ns = strings.Trim(string(envFileContent), " \n")
				}
			}
		} else {
			ns = DEFAULT_NAMESPACE
		}
	} else {
		ns = os.Getenv("t")
	}

	namespacePath := path.Join(home, T_BASE_DIR, ns)

	fstat, err := os.Stat(namespacePath)
	if err != nil {
		mkdirError := os.MkdirAll(namespacePath, 0755)
		if mkdirError != nil {
			die("Cant create namespace: %s", mkdirError)
		}
	} else {
		if !fstat.IsDir() {
			die("Selected namespace not a directory")
		}
	}

	tasks, err := getTasksInNamespaceSorted(namespacePath)
	if err != nil {
		die("Error get tasks: %s", err)
	}

	if len(os.Args) < 2 {
		showTasks(tasks, path.Join(home, T_BASE_DIR, ns))
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "show":
		showTasks(tasks, path.Join(home, T_BASE_DIR, ns))
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "a", "add":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		addTask(path.Join(home, T_BASE_DIR, ns), os.Args[2:])
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "d", "done", "delete":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := deleteTasksByIndexes(os.Args[2:], tasks, path.Join(home, T_BASE_DIR, ns))
		if err != nil {
			die("Error deleting task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "e", "edit":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := editTaskByIndex(os.Args[2], tasks, path.Join(home, T_BASE_DIR, ns))
		if err != nil {
			die("Error editing task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "get":
		if len(os.Args) < 3 {
			die("Not enough args")
		}

		err := showTaskContentByName(path.Join(home, T_BASE_DIR, ns), os.Args[2])
		if err != nil {
			die("Error reading task: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "ns", "namespaces":
		err := showNamespaces(path.Join(home, T_BASE_DIR))
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
		err := showTaskContentByIndex(cmd, tasks, path.Join(home, T_BASE_DIR, ns))
		if err != nil {
			die("Error: %s", err)
		}

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}
}

func showTasks(tasks []string, namespace string) {
	fmt.Printf("\033[1;34m# %s\033[0m\n", path.Base(namespace))
	for i, task := range tasks {
		taskLines, _ := countFileLines(path.Join(namespace, task))
		var formattedTaskLines string

		if taskLines > 70 {
			formattedTaskLines = "..."
		} else if taskLines == 0 {
			formattedTaskLines = "-"
		} else {
			formattedTaskLines = fmt.Sprint(taskLines + 1)
		}

		formattedTaskName := strings.ReplaceAll(task, PATH_SEPARATOR_REPLACER, "/")
		fmt.Printf("[%d] %s (%s)\n", i+1, formattedTaskName, formattedTaskLines)
	}
}

func addTask(namespace string, taskName []string) {
	newTaskName := strings.Join(taskName, " ")
	newTaskName = strings.ReplaceAll(newTaskName, "/", PATH_SEPARATOR_REPLACER)

	err := os.WriteFile(path.Join(namespace, newTaskName), []byte{}, 0644)
	if err != nil {
		die("Error write task: %s", err)
	}
}

func deleteTasksByIndexes(indexes []string, tasks []string, namespace string) error {
	for _, inputedTaskIndex := range indexes {
		taskIndex, err := strconv.Atoi(inputedTaskIndex)
		if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
			return fmt.Errorf("Wrong task index: %s", inputedTaskIndex)
		}

		taskNameToDelete := tasks[taskIndex-1]
		deleteErr := os.Remove(path.Join(namespace, taskNameToDelete))
		if deleteErr != nil {
			return fmt.Errorf("Error remove task: %s", deleteErr)
		}
	}

	return nil
}

func editTaskByIndex(index string, tasks []string, namespace string) error {
	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return fmt.Errorf("Wrong task index")
	}

	taskIndexToEdit := tasks[taskIndex-1]
	taskToEdit := path.Join(namespace, taskIndexToEdit)

	cmd := exec.Command(os.Getenv("EDITOR"), taskToEdit)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error run EDITOR: %w", err)
	}

	return nil
}

func showTaskContentByName(namespace string, name string) error {
	content, err := os.ReadFile(path.Join(namespace, name))
	if err != nil {
		return fmt.Errorf("Error reading task: %w", err)
	}

	fmt.Print(string(content))
	return nil
}

func showNamespaces(tBaseDir string) error {
	dirEntries, err := os.ReadDir(tBaseDir)
	if err != nil {
		return err
	}

	for _, de := range dirEntries {
		if de.Name()[0] == '.' {
			continue
		}
		if de.IsDir() {
			namespaceDirEntries, err := os.ReadDir(path.Join(tBaseDir, de.Name()))
			namespaceTasksCount := 0
			if err == nil {
				namespaceTasksCount = len(namespaceDirEntries)
			}
			fmt.Printf("%s (%d)\n", de.Name(), namespaceTasksCount)
		}
	}
	return nil
}

func showHelp() {
	fmt.Print(HELP_MESSAGE)
}

func showVersion() {
	fmt.Print(version)
}

func showTaskContentByIndex(cmd string, tasks []string, namespace string) error {
	taskIndex, err := strconv.Atoi(cmd)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return fmt.Errorf("Wrong task index: %s", cmd)
	}

	taskNameToRead := tasks[taskIndex-1]
	taskContent, err := os.ReadFile(path.Join(namespace, taskNameToRead))
	if err != nil {
		return fmt.Errorf("Error reading task: %w", err)
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n\n", taskNameToRead)
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

func getTasksInNamespaceSorted(namespacePath string) ([]string, error) {
	dirEntries, err := os.ReadDir(namespacePath)
	if err != nil {
		return nil, err
	}

	sortErr := sortTasks(dirEntries)
	if sortErr != nil {
		die("Error sorting tasks: %s", sortErr)
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

func countFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1024)
	count := 0
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
