package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
)

const T_BASE_DIR = ".t"
const DEFAULT_NAMESPACE = "def"

func main() {
	home := os.Getenv("HOME")

	ns := os.Getenv("t")
	if ns == "" {
		ns = DEFAULT_NAMESPACE
	}

	namespacePath := path.Join(home, T_BASE_DIR, ns)

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

			var noteLinesFormatted string

			if noteLines > 70 {
				noteLinesFormatted = "..."
			} else if noteLines == 0 {
				noteLinesFormatted = "-"
			} else {
				noteLinesFormatted = fmt.Sprint(noteLines + 1)
			}

			fmt.Printf("[%d] %s (%s)\n", i+1, note, noteLinesFormatted)
		}
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
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
	}
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