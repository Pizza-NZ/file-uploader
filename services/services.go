package services

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/pizza-nz/file-uploader/storage"
	"github.com/pizza-nz/file-uploader/types"
)

type FileUploadService interface {
	CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

type FileUploadServiceImpl struct {
	fileStorage storage.FileStorage
}

func NewFileUploadService(fileStorage storage.FileStorage) FileUploadService {
	return &FileUploadServiceImpl{fileStorage: fileStorage}
}

func (s *FileUploadServiceImpl) CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
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
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
	if !allowedTypes[contentType] {
		return nil, types.NewAppError("Invalid File Type", fmt.Sprintf("File type %s is not allowed", contentType), http.StatusBadRequest, nil)
	}

	s3ObjectKey, err := s.fileStorage.Upload(ctx, file, handler)
	if err != nil {
		return nil, err
	}

	slog.Info("File uploaded successfully", "filename", handler.Filename, "s3_key", s3ObjectKey)
	return &types.FileUploadResponse{FileID: s3ObjectKey, Size: handler.Size}, nil
}
