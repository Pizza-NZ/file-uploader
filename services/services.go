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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	ctx, span := otel.Tracer("file-uploader/service").Start(ctx, "CreateFileUpload")
	defer span.End()

	// Read the first 261 bytes to determine the file type
	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil && err != io.EOF {
		span.SetAttributes(attribute.String("create_file_upload.read", "failed to read file header"))
		span.AddEvent("File header read failed")
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset the file reader so the full file can be read again later
	if _, err := file.Seek(0, 0); err != nil {
		span.SetAttributes(attribute.String("create_file_upload.read", "failed to reset file header"))
		span.AddEvent("File header reset failed")
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Use filetype.Match to determine the file type based on magic numbers
	kind, err := filetype.Match(head)
	if err != nil {
		span.SetAttributes(attribute.String("create_file_upload.match", "failed to match file type"))
		span.AddEvent("File type match failed")
		return nil, fmt.Errorf("failed to match file type: %w", err)
	}

	// Check if the detected file type is allowed
	if kind == filetype.Unknown || !s.allowedTypes[kind.MIME.Value] {
		span.SetAttributes(attribute.String("create_file_upload.type", "invalid file type"))
		span.AddEvent("Invalid file type")
		return nil, types.NewAppError("Invalid File Type", fmt.Sprintf("File type %s is not allowed", kind.MIME.Value), http.StatusBadRequest, nil)
	}

	objectKey, err := s.fileStorage.Upload(ctx, file, handler)
	if err != nil {
		return nil, err
	}

	slog.Info("File uploaded successfully", "filename", handler.Filename, "object_key", objectKey)
	return &types.FileUploadResponse{FileID: objectKey, Size: handler.Size}, nil
}
