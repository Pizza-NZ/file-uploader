package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/h2non/filetype"
	"github.com/pizza-nz/file-uploader/storage"
	"github.com/pizza-nz/file-uploader/types"
)

type FileUploadService interface {
	CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

type FileUploadServiceImpl struct {
	fileStorage  storage.FileStorage
	allowedTypes map[string]bool
}

func NewFileUploadService(fileStorage storage.FileStorage, allowedTypes []string) FileUploadService {
	allowedTypesMap := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedTypesMap[t] = true
	}
	return &FileUploadServiceImpl{fileStorage: fileStorage, allowedTypes: allowedTypesMap}
}

func (s *FileUploadServiceImpl) CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	defer file.Close()

	// Read the first 261 bytes to determine the file type
	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset the file reader so the full file can be read again later
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Use filetype.Match to determine the file type based on magic numbers
	kind, err := filetype.Match(head)
	if err != nil {
		return nil, fmt.Errorf("failed to match file type: %w", err)
	}

	// Check if the detected file type is allowed
	if kind == filetype.Unknown || !s.allowedTypes[kind.MIME.Value] {
		return nil, types.NewAppError("Invalid File Type", fmt.Sprintf("File type %s is not allowed", kind.MIME.Value), http.StatusBadRequest, nil)
	}

	s3ObjectKey, err := s.fileStorage.Upload(ctx, file, handler)
	if err != nil {
		return nil, err
	}

	slog.Info("File uploaded successfully", "filename", handler.Filename, "s3_key", s3ObjectKey)
	return &types.FileUploadResponse{FileID: s3ObjectKey, Size: handler.Size}, nil
}
