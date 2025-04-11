//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	versionFileContent, readErr := os.ReadFile("../../VERSION")
	if readErr != nil {
		log.Fatalf("Failed to read version file: %s", readErr)
	}

	version := strings.Trim(string(versionFileContent), " \n")
	content := fmt.Sprintf(`//go:build !tsqlite

package main

var (
	version = "%s"
)
`, version)

	err := os.WriteFile("version_info.go", []byte(content), 0644)
	if err != nil {
		log.Fatalf("Failed to write version info: %s", err)
	}

	contentSqlite := fmt.Sprintf(`//go:build tsqlite

package main

var (
	version = "%s-sqlite"
)
`, version)
	err = os.WriteFile("version_info_sqlite.go", []byte(contentSqlite), 0644)
	if err != nil {
		log.Fatalf("Failed to write version info: %s", err)
	}
}
