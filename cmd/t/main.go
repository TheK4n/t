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
const HELP_MESSAGE = `USAGE
    T script for fast notes

    t                            - Show notes in format '[INDEX] NOTE NAME (LINES)'
    t get (NOTE)                 - Get note content
    t show                       - Show notes in format '[INDEX] NOTE NAME (LINES)'
    t (INDEX)                    - Show note content
    t add (X X X)                - Add note with name X X X
    t edit (INDEX)               - Edit note with INDEX by \$EDITOR
    t done (INDEX) [INDEX] ...   - Delete notes with INDEXes
    t namespaces                 - Show namespaces
    t --help                     - Show this message

    t a       - alias for add
    t e       - alias for edit
    t d       - alias for done
    t delete  - alias for done
    t ns      - alias for namespaces


NAMESPACES
    t namespaces             # show namespaces
    t=work t a fix bug 211   # add note in workspace 'work'
    t=work t                 # show notes in workspace 'work'`

func main() {
	home := os.Getenv("HOME")

	var ns string

	if os.Getenv("t") == "" {
		curdir, _ := os.Getwd()

		foundEnvFile := findFileAbsPathUpTree(curdir, ENVFILE)
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
			panic("Cant create namespace")
		}
	} else {
		if !fstat.IsDir() {
			panic("Selected namespace not a directory")
		}
	}

	notes, err := getNotesInDirSorted(namespacePath)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		fmt.Printf("\033[1;34m# %s\033[0m\n", ns)
		for i, note := range notes {
			noteLines, err := countFileLines(path.Join(home, T_BASE_DIR, ns, note))
			if err != nil {
				panic(err)
			}

			var formattedNoteLines string

			if noteLines > 70 {
				formattedNoteLines = "..."
			} else if noteLines == 0 {
				formattedNoteLines = "-"
			} else {
				formattedNoteLines = fmt.Sprint(noteLines + 1)
			}

			formattedNoteName := strings.ReplaceAll(note, PATH_SEPARATOR_REPLACER, "/")
			fmt.Printf("[%d] %s (%s)\n", i+1, formattedNoteName, formattedNoteLines)
		}
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "a", "add":
		if len(os.Args) < 3 {
			panic("not enough args")
		}

		newNoteName := strings.Join(os.Args[2:], " ")
		newNoteName = strings.ReplaceAll(newNoteName, "/", PATH_SEPARATOR_REPLACER)

		err := os.WriteFile(path.Join(home, T_BASE_DIR, ns, newNoteName), []byte{}, 0644)
		if err != nil {
			panic(err)
		}
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "d", "done", "delete":
		if len(os.Args) < 3 {
			panic("not enougn args")
		}

		for _, inputedNoteIndex := range os.Args[2:] {
			noteIndex, err := strconv.Atoi(inputedNoteIndex)
			if err != nil || noteIndex > len(notes) || noteIndex < 1 {
				panic("wrong note index")
			}

			noteToRemove := notes[noteIndex-1]
			removeErr := os.Remove(path.Join(home, T_BASE_DIR, ns, noteToRemove))
			if removeErr != nil {
				panic(removeErr)
			}
		}
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "e", "edit":
		if len(os.Args) < 3 {
			panic("not enougn args")
		}

		noteIndex, err := strconv.Atoi(os.Args[2])
		if err != nil || noteIndex > len(notes) || noteIndex < 1 {
			panic("wrong note index")
		}

		noteIndexToEdit := notes[noteIndex-1]
		noteToEdit := path.Join(home, T_BASE_DIR, ns, noteIndexToEdit)

		cmd := exec.Command(os.Getenv("EDITOR"), noteToEdit)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			panic(err)
		}
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "get":
		if len(os.Args) < 3 {
			panic("not enougn args")
		}

		content, err := os.ReadFile(path.Join(home, T_BASE_DIR, ns, os.Args[2]))
		if err != nil {
			panic(err)
		}
		fmt.Print(string(content))
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "ns", "namespaces":
		dirEntries, err := os.ReadDir(path.Join(home, T_BASE_DIR))
		if err != nil {
			panic(err)
		}

		for _, de := range dirEntries {
			if de.Name()[0] == '.' {
				continue
			}
			if de.IsDir() {
				namespaceDirEntries, err := os.ReadDir(path.Join(home, T_BASE_DIR, de.Name()))
				namespaceNotesCount := 0
				if err == nil {
					namespaceNotesCount = len(namespaceDirEntries)
				}
				fmt.Printf("%s (%d)\n", de.Name(), namespaceNotesCount)
			}
		}
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	case "-h", "--help":
		fmt.Print(HELP_MESSAGE)

		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)

	default:
		noteIndex, err := strconv.Atoi(cmd)
		if err != nil || noteIndex > len(notes) || noteIndex < 1 {
			panic("wrong note index")
		}

		noteToRead := notes[noteIndex-1]
		fmt.Printf("\033[1;34m# %s\033[0m\n\n", noteToRead)

		noteContent, err := os.ReadFile(path.Join(home, T_BASE_DIR, ns, noteToRead))
		if err != nil {
			panic(err)
		}
		fmt.Print(string(noteContent))
		removeEmptyNamespaces(path.Join(home, T_BASE_DIR))
		os.Exit(0)
	}
}

func findFileAbsPathUpTree(startdir string, filename string) string {
	if startdir == "/" {
		return ""
	}
	if _, err := os.Stat(path.Join(startdir, filename)); err == nil {
		return path.Join(startdir, filename)
	}
	return findFileAbsPathUpTree(filepath.Dir(startdir), filename)
}

func getNotesInDirSorted(namespacePath string) ([]string, error) {
	dirEntries, err := os.ReadDir(namespacePath)
	if err != nil {
		return nil, err
	}

	sortErr := sortNotes(dirEntries)
	if sortErr != nil {
		panic(sortErr)
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

func sortNotes(dirEntries []os.DirEntry) error {
	var sortErr error

	sort.Slice(dirEntries, func(i, j int) bool {
		iInfo, err := dirEntries[i].Info()
		jInfo, err := dirEntries[j].Info()

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

func removeEmptyNamespaces(dir string) error {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, de := range dirEntries {
		subdirEntries, err := os.ReadDir(path.Join(dir, de.Name()))
		if err != nil {
			continue
		}
		if len(subdirEntries) < 1 {
			rmErr := os.Remove(path.Join(dir, de.Name()))
			if rmErr != nil {
				panic(rmErr)
			}
		}
	}

	return nil
}
