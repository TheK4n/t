package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	storage "github.com/thek4n/t/internal/storage"
)

const PATH_SEPARATOR_REPLACER = "%2F"
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

	t <namespace> ...        # optional argument namespace before commands

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

type TaskView struct {
	Namespace           string
	LinesCount          int
	FormattedLinesCount string
	Name                string
	FormattedName       string
}

func ShowTasks(namespace string, s storage.TasksStorage) error {
	tasks, err := s.GetSorted(namespace)
	if err != nil {
		return err
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n", namespace)
	for i, task := range tasks {
		tv := formatTaskView(namespace, task, s)
		fmt.Printf("[%d] %s (%s)\n", i+1, tv.FormattedName, tv.FormattedLinesCount)
	}

	return nil
}

func formatLinesCount(lines int) string {
	if lines > 70 {
		return "..."
	}
	if lines == 0 {
		return "-"
	}
	return fmt.Sprint(lines + 1)
}

func AddTask(namespace string, name string, s storage.TasksStorage) error {
	return s.Add(namespace, name)
}

func DeleteTasksByIndexes(namespace string, indexes []int, s storage.TasksStorage) error {
	return s.DeleteByIndexes(namespace, indexes)
}

func EditTaskByIndex(namespace string, index int, s storage.TasksStorage) error {
	taskName, err := s.GetNameByIndex(namespace, index)
	if err != nil {
		return err
	}

	tempFile, err := createTempFile(fmt.Sprintf("t_%s_", taskName))
	if err != nil {
		return err
	}

	content, err := s.GetContentByIndex(namespace, index)
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

func ShowTaskContentByName(namespace string, name string, s storage.TasksStorage) error {
	content, err := s.GetContentByName(namespace, name)
	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

func ShowTaskContentByIndex(namespace string, index int, s storage.TasksStorage) error {
	taskContent, err := s.GetContentByIndex(namespace, index)
	taskName, err := s.GetNameByIndex(namespace, index)

	if err != nil {
		return err
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n\n", taskName)
	fmt.Print(string(taskContent))
	return nil
}

func ShowNamespaces(s storage.TasksStorage) error {
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

func ShowHelp() error {
	_, err := fmt.Print(HELP_MESSAGE)
	return err
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
			tv := formatTaskView(namespace, task, s)
			fmt.Printf("[%s] %s (%s)\n", tv.Namespace, tv.FormattedName, tv.FormattedLinesCount)
		}
	}
	return nil
}

func formatTaskView(namespace string, task string, s storage.TasksStorage) TaskView {
	var tv TaskView

	lines, err := s.CountLines(namespace, task)
	if err != nil {
		lines = 0
	}
	tv.LinesCount = lines
	tv.FormattedLinesCount = formatLinesCount(lines)
	tv.Name = task
	tv.Namespace = namespace
	tv.FormattedName = strings.ReplaceAll(task, PATH_SEPARATOR_REPLACER, "/")

	return tv
}
