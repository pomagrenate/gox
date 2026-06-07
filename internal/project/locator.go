package project

import (
	"errors"
	"os"
	"path/filepath"
)

func FindProjectRoot(startDir string) (string, error) {
	currentDir := startDir

	for {
		goxYamlPath := filepath.Join(currentDir, "gox.yaml")
		if _, err := os.Stat(goxYamlPath); err == nil {
			return currentDir, nil
		}

		goWorkPath := filepath.Join(currentDir, "go.work")
		if _, err := os.Stat(goWorkPath); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	return "", errors.New("Can not find gox.yaml or go.work in project tree")
}
