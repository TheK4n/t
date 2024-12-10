package main

import (
	"fmt"
	"os"
	"path"
	"sort"
)


const T_BASE_DIR = ".t"
const DEFAULT_NAMESPACE = "def"


func main() {
	home := os.Getenv("HOME")

	ns := os.Getenv("t")
	if ns == "" {
		ns = DEFAULT_NAMESPACE
	}

	fmt.Printf("\033[1;34m# %s\033[0m\n", ns)

	namespacePath := path.Join(home, T_BASE_DIR, ns)

	notes, err := getNotesInDirSorted(namespacePath)
	if err != nil {
		panic(err)
	}

	for i, note := range notes {
		fmt.Printf("[%d] %s\n", i+1, note)
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