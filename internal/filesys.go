package internal

import (
	"fmt"
	"log"
	"os"
)

func getPWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get current directory: %w", err))
	}
	return pwd
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
