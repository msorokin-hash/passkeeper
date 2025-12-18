package utils

import (
	"errors"
	"fmt"
	"os"
)

// ReadFile reads the contents of the specified file path and returns it as a string.
// It validates that the path exists, is not a directory, and is readable.
// Returns an error if the file doesn't exist, is a directory, or cannot be read.
func ReadFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file does not exist: " + path)
		}
		return "", err
	}

	if info.IsDir() {
		return "", errors.New("path is a directory, not a file: " + path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// AppendToFile appends text to the end of a file.
// Creates the file with 0666 permissions if it doesn't exist.
// Returns an error if file operations fail.
func AppendToFile(path string, text string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove existing file: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}
