package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// ReadFile reads the entire contents of the file at the given path
// and returns it as a string.
//
// The function attempts to read the file directly and relies on the
// underlying file system errors to determine failure conditions.
// It returns a wrapped error if the file does not exist, access is denied,
// or the file cannot be read for any other reason.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("file does not exist: %s", path)
		}
		if errors.Is(err, fs.ErrPermission) {
			return "", fmt.Errorf("permission denied: %s", path)
		}
		return "", err
	}

	return string(data), nil
}

// OverwriteFile truncates the file at the given path and writes the provided text.
//
// If the file already exists, its contents are removed before writing.
// If the file does not exist, it is created with 0666 permissions
// (before applying the process umask).
func OverwriteFile(path string, text string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}
