package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "hello world"
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	readContent, err := ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestReadFile_NotExists(t *testing.T) {
	_, err := ReadFile("no_such_file_12345.txt")
	assert.Error(t, err)
}

func TestAppendToFile_CreateIfNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new_file.txt")

	text := "hello world\n"

	err := AppendToFile(path, text)
	assert.NoError(t, err)

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Equal(t, text, string(data))
}

func TestAppendToFile_OverrideIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.txt")

	initial := "old text\n"
	err := os.WriteFile(path, []byte(initial), 0644)
	assert.NoError(t, err)

	newText := "new content\n"
	err = AppendToFile(path, newText)
	assert.NoError(t, err)

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Equal(t, newText, string(data))
}
