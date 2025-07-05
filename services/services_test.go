package services

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mocking multipart.File
type mockMultipartFile struct {
	*bytes.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func (m *mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.Reader.ReadAt(p, off)
}

func TestCreateFileUpload_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-temp-files")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the RootPath to the temporary directory
	originalRootPath := RootPath
	RootPath = tempDir
	defer func() { RootPath = originalRootPath }()

	// Create the tempFiles directory
	os.Mkdir(filepath.Join(RootPath, "tempFiles"), 0755)

	// Create a dummy file content
	fileContent := []byte("test file content")
	file := &mockMultipartFile{bytes.NewReader(fileContent)}
	handler := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     int64(len(fileContent)),
	}

	service := &FileUploadServiceImpl{}
	response, err := service.CreateFileUpload(file, handler)
	assert.NoError(t, err)

	// Check if the file was created in the tempFiles directory
	files, err := os.ReadDir(filepath.Join(RootPath, "tempFiles"))
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.True(t, strings.HasPrefix(files[0].Name(), "upload-test-"))
	assert.True(t, strings.HasSuffix(files[0].Name(), ".txt"))
	assert.Equal(t, int64(len(fileContent)), response.Size)
	assert.True(t, strings.HasPrefix(response.FileID, "upload-test-"))
	assert.True(t, strings.HasSuffix(response.FileID, ".txt"))
}

func TestCreateFileUpload_ErrorCreatingFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-temp-files")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the RootPath to the temporary directory
	originalRootPath := RootPath
	RootPath = tempDir
	defer func() { RootPath = originalRootPath }()

	// Change permissions to read-only to simulate an error creating a file
	err = os.Chmod(tempDir, 0444)
	assert.NoError(t, err)

	// Create a dummy file content
	fileContent := []byte("test file content")
	file := &mockMultipartFile{bytes.NewReader(fileContent)}
	handler := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     int64(len(fileContent)),
	}

	service := NewFileUploadService(tempDir)
	_, err = service.CreateFileUpload(file, handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
