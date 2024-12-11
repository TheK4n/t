//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	versionFileContent, readErr := os.ReadFile("../../VERSION")
	if readErr != nil {
		log.Fatalf("Failed to read version file: %s", readErr)
	}

	version := string(versionFileContent)
	content := fmt.Sprintf(`package main

var (
    version   = "%s"
)
`, version)

	err := os.WriteFile("version_info.go", []byte(content), 0666)
	if err != nil {
		log.Fatalf("Failed to write version info: %s", err)
	}
}
