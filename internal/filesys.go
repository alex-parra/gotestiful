package internal

import (
	"fmt"
	"os"
)

func getPWD() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return pwd, err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readFile(filePath string) ([]byte, error) {
	fileExists := fileExists(filePath)
	if !fileExists {
		return nil, fmt.Errorf("file not found %s", filePath)
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fileBytes, nil
}
