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
	"time"

	handlers "github.com/thek4n/t/internal/handlers"
	"github.com/thek4n/t/internal/storage"
)

const DEFAULT_NAMESPACE = "def"
const ENVFILE = ".tns"

var COMMANDS = map[string]func(storage.TasksStorage, []string, string) error{
	"show": cmdShow,

	"add": cmdAdd,
	"a":   cmdAdd,
	"ф":   cmdAdd,

	"done":   cmdDone,
	"delete": cmdDone,
	"d":      cmdDone,
	"в":      cmdDone,

	"edit": cmdEdit,
	"у":    cmdEdit,

	"get": cmdGet,

	"ns":         cmdNamespaces,
	"namespaces": cmdNamespaces,

	"all": cmdAll,

	"deadline": cmdSetDeadline,
	"dl":       cmdSetDeadline,

	"-h":     cmdHelp,
	"--help": cmdHelp,

	"-v":        cmdVersion,
	"--version": cmdVersion,
}

func main() {
	osArgs := os.Args[1:] // reject program name

	s := initTaskStorage()

	err := notifyExpired(s)
	if err != nil {
		die("Error on notify")
	}

	argsEmpty := len(osArgs) < 1
	if argsEmpty {
		namespace := getNamespace()
		err := showTasks(s, namespace)
		if err != nil {
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
			die("Error show namespaces: %s", err)
		}
		os.Exit(0)
	}

	commandArgumentIsNumber, _ := regexp.MatchString(`[0-9]+`, osArgs[0])
	if commandArgumentIsNumber {
		index, err := strconv.Atoi(osArgs[0])
		if err != nil {
			die("Error parse index")
		}

		err = handlers.ShowTaskContentByIndex(namespace, index, s)
		if err != nil {
			die("Error: %s", err)
		}

		os.Exit(0)
	}

	handler, found := COMMANDS[osArgs[0]]
	if !found {
		die("Command '%s' not found", osArgs[0])
	}

	err = handler(s, osArgs[1:], namespace)
	if err != nil {
		die("Error on command '%s': %s", osArgs[0], err)
	}

	os.Exit(0)
}

func notifyExpired(s storage.TasksStorage) error {
	tasks, err := s.GetExpired()
	if err != nil {
		return err
	}
	if len(tasks) < 1 {
		return nil
	}

	fmt.Printf("\033[1;33m# Notifications!\033[0m\n")

	for _, task := range tasks {
		fmt.Printf("[%s] %s\n", task.Namespace, task.Name)
	}

	fmt.Printf("\n")
	return nil
}

func showTasks(s storage.TasksStorage, namespace string) error {
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

func cmdSetDeadline(s storage.TasksStorage, args []string, namespace string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s", "Not enough args")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("Error parse index %s: %s", args[0], err)
	}

	name, err := s.GetNameByIndex(namespace, index)
	if err != nil {
		return fmt.Errorf("Error setting deadline for index %s: %s", args[0], err)
	}

	date, err := parseTime(args[1])
	if err != nil {
		return fmt.Errorf("Error parse date '%s': %s", args[1], err)
	}

	return s.SetNotifyDeadline(namespace, name, date)
}

func parseTime(t string) (time.Time, error) {
	reTime := regexp.MustCompile(`^\d{1,2}:\d{2}$`)
	reDay := regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{2}$`)

	switch {
	case reTime.Match([]byte(t)):
		return timeCurrentDayOrNextDay(t)
	case reDay.Match([]byte(t)):
		return specifiedDayThatTime(t)
	}

	switch t {
	case "tommorow":
		return nextDayThatTime(), nil
	case "morning":
		return nextDayMorning(), nil
	case "week":
		return nextWeekMorning(), nil
	}

	return time.Time{}, fmt.Errorf("No match")
}

func timeCurrentDayOrNextDay(t string) (time.Time, error) {
	datetime, err := time.Parse("15:04", t)
	if err != nil {
		return datetime, err
	}

	now := time.Now()
	nowDatePlusSpecifiedTime := time.Date(now.Year(), now.Month(), now.Day(), datetime.Hour(), datetime.Minute(), datetime.Second(), 0, now.Location())
	if now.Compare(nowDatePlusSpecifiedTime) == 1 {
		nowDatePlusSpecifiedTime = nowDatePlusSpecifiedTime.Add(24 * time.Hour)
	}

	return nowDatePlusSpecifiedTime, nil
}

func specifiedDayThatTime(t string) (time.Time, error) {
	now := time.Now()
	day, err := time.Parse("2.1.06", t)
	if err != nil {
		return day, err
	}
	return time.Date(day.Year(), day.Month(), day.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location()), nil
}

func nextDayMorning() time.Time {
	now := time.Now()
	yyyy, mm, dd := now.Date()
	return time.Date(yyyy, mm, dd+1, 8, 0, 0, 0, now.Location())
}

func nextWeekMorning() time.Time {
	now := time.Now()
	yyyy, mm, dd := now.Date()
	return time.Date(yyyy, mm, dd+7, 8, 0, 0, 0, now.Location())
}

func nextDayThatTime() time.Time {
	return time.Now().Add(time.Hour * 24)
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
