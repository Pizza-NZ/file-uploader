package services

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/pizza-nz/file-uploader/types"
	"github.com/pizza-nz/file-uploader/utils"
)

var (
	_, b, _, _ = runtime.Caller(0)
	RootPath   = filepath.Join(filepath.Dir(b), "../..")
	createDirOnce sync.Once
)

type FileUploadService interface {
	CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

type FileUploadServiceImpl struct {
	filePath string
}

func NewFileUploadService(filePath string) FileUploadService {
	return &FileUploadServiceImpl{filePath: filePath}
}

func (s *FileUploadServiceImpl) CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	defer file.Close()

	// Check file type
	fileHeader := make([]byte, 512)
	if _, err := file.Read(fileHeader); err != nil {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}
	contentType := http.DetectContentType(fileHeader)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"application/pdf": true,
	}
	if !allowedTypes[contentType] {
		return nil, types.NewAppError("Invalid File Type", fmt.Sprintf("File type %s is not allowed", contentType), http.StatusBadRequest, nil)
	}

	tempFolderPath := s.filePath
	createDirOnce.Do(func() {
		if _, err := os.Stat(tempFolderPath); os.IsNotExist(err) {
			if err := os.MkdirAll(tempFolderPath, os.ModePerm); err != nil {
				slog.Error("Error creating temporary folder", "error", err)
			}
		}
	})

	slog.Info("Creating temporary folder", "path", tempFolderPath)
	tempFileName := fmt.Sprintf("upload-%s-*%s", utils.FileNameWithoutExtension(handler.Filename), filepath.Ext(handler.Filename))

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	if err != nil {
		slog.Error("Error creating temporary file", "error", err)
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer tempFile.Close()

	filebytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading file buffer", "error", err)
		return nil, fmt.Errorf("failed to read file buffer: %w", err)
	}

	_, err = tempFile.Write(filebytes)
	if err != nil {
		slog.Error("Error writing file to disk", "error", err)
		return nil, fmt.Errorf("failed to write file to disk: %w", err)
	}

	slog.Info("File uploaded successfully", "filename", handler.Filename)
	return &types.FileUploadResponse{FileID: filepath.Base(tempFile.Name()), Size: handler.Size}, nil
}
