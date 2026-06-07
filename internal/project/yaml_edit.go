package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InsertIntoYaml finds a specific block (like "apps:" or "libs:") and inserts the payload below it.
func InsertIntoYaml(filePath string, section string, payload string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inserted := false

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		if strings.TrimSpace(line) == section && !inserted {
			lines = append(lines, payload)
			inserted = true
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !inserted {
		return fmt.Errorf("section '%s' not found in %s", section, filePath)
	}

	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filePath, []byte(output), 0644)
}
