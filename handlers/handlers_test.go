
package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pizza-nz/file-uploader/types"
	"github.com/stretchr/testify/assert"
)

// Mock FileUploadService
type MockFileUploadService struct {
	CreateFileUploadFunc func(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

func (m *MockFileUploadService) CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	return m.CreateFileUploadFunc(file, handler)
}

func TestCreateFileUpload(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.WriteString("test file content")
	tempFile.Close()

	// Re-open the file for reading
	file, err := os.Open(tempFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Create a buffer to store the multipart form data
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Create a form file
	formFile, err := multipartWriter.CreateFormFile("uploadFile", filepath.Base(tempFile.Name()))
	assert.NoError(t, err)

	// Copy the file content to the form file
	_, err = io.Copy(formFile, file)
	assert.NoError(t, err)

	// Close the multipart writer
	multipartWriter.Close()

	// Test cases
	tests := []struct {
		name               string
		maxFileSize        int64
		service            *MockFileUploadService
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:        "Successful file upload",
			maxFileSize: 10 * 1024 * 1024, // 10 MB
			service: &MockFileUploadService{
				CreateFileUploadFunc: func(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
					return &types.FileUploadResponse{FileID: "test-file-id", Size: 123}, nil
				},
			},
			expectedStatusCode: http.StatusCreated,
			expectedBody:       `"fileId":"test-file-id","size":123`,
		},
		{
			name:               "No file in upload",
			maxFileSize:        10 * 1024 * 1024, // 10 MB
			service:            &MockFileUploadService{},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `"Error Reading File"`,
		},
		{
			name:        "File greater than maxFileSize",
			maxFileSize: 5, // 5 bytes
			service:     &MockFileUploadService{},
			// The error for file size is handled by r.ParseMultipartForm, so the handler won't even be called.
			// We expect a bad request.
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `"Error Reading File"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/upload", &requestBody)
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			if tt.name == "No file in upload" {
				// Create a new request with an empty body
				req = httptest.NewRequest("POST", "/upload", nil)
				req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
			}

			w := httptest.NewRecorder()

			handler := &FileUploadHandlerImpl{
				maxFileSize: tt.maxFileSize,
				service:     tt.service,
			}

			handler.CreateFileUpload(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
