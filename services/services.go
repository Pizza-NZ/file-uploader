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
	"go.opentelemetry.io/otel/metric"
)

type FileUploadService interface {
	CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

type FileUploadServiceImpl struct {
	fileStorage          storage.FileStorage
	allowedTypes         map[string]bool
	meter                metric.Meter
	uploadCounter        metric.Int64Counter
	uploadSize           metric.Int64Histogram
	rejectionCounter     metric.Int64Counter
	uploadFailureCounter metric.Int64Counter
}

func NewFileUploadService(fileStorage storage.FileStorage, allowedTypes []string) FileUploadService {
	allowedTypesMap := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedTypesMap[t] = true
	}
	meter := otel.GetMeterProvider().Meter("file-uploader/service")
	uploadCounter, _ := meter.Int64Counter("file.uploads.total", metric.WithDescription("Total number of file uploads."))
	uploadSize, _ := meter.Int64Histogram("file.upload.size", metric.WithUnit("By"), metric.WithDescription("Size of uploaded files in bytes."))
	rejectionCounter, _ := meter.Int64Counter("file.rejections.total", metric.WithDescription("Total number of rejected files."))
	uploadFailureCounter, _ := meter.Int64Counter("file.upload.failures.total", metric.WithDescription("Total number of failed file uploads."))
	return &FileUploadServiceImpl{
		fileStorage:          fileStorage,
		allowedTypes:         allowedTypesMap,
		meter:                meter,
		uploadCounter:        uploadCounter,
		uploadSize:           uploadSize,
		rejectionCounter:     rejectionCounter,
		uploadFailureCounter: uploadFailureCounter,
	}
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

		rejectionReason := attribute.String("reason", "file_header_read_error")
		s.rejectionCounter.Add(ctx, 1, metric.WithAttributes(rejectionReason))
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset the file reader so the full file can be read again later
	if _, err := file.Seek(0, 0); err != nil {
		span.SetAttributes(attribute.String("create_file_upload.read", "failed to reset file header"))
		span.AddEvent("File header reset failed")

		rejectionReason := attribute.String("reason", "file_header_reset_error")
		s.rejectionCounter.Add(ctx, 1, metric.WithAttributes(rejectionReason))
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Use filetype.Match to determine the file type based on magic numbers
	kind, err := filetype.Match(head)
	if err != nil {
		span.SetAttributes(attribute.String("create_file_upload.match", "failed to match file type"))
		span.AddEvent("File type match failed")

		rejectionReason := attribute.String("reason", "filetype_match_error")
		s.rejectionCounter.Add(ctx, 1, metric.WithAttributes(rejectionReason))
		return nil, fmt.Errorf("failed to match file type: %w", err)
	}

	// Check if the detected file type is allowed
	if kind == filetype.Unknown || !s.allowedTypes[kind.MIME.Value] {
		err := fmt.Errorf("File type %s is not allowed", kind.MIME.Value)

		span.SetAttributes(attribute.String("create_file_upload.type", "invalid file type"))
		span.AddEvent("Invalid file type")

		rejectionReason := attribute.String("reason", "invalid_type "+kind.MIME.Value)
		s.rejectionCounter.Add(ctx, 1, metric.WithAttributes(rejectionReason))
		return nil, types.NewAppError("Invalid File Type", err.Error(), http.StatusBadRequest, err)
	}

	objectKey, err := s.fileStorage.Upload(ctx, file, handler)
	if err != nil {

		rejectionReason := attribute.String("reason", "storage_error "+err.Error())
		s.uploadFailureCounter.Add(ctx, 1, metric.WithAttributes(rejectionReason))
		return nil, err
	}

	slog.InfoContext(ctx, "File uploaded successfully", "filename", handler.Filename, "object_key", objectKey)
	s.uploadCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("file.name", handler.Filename)))
	s.uploadSize.Record(ctx, handler.Size, metric.WithAttributes(attribute.String("file.name", handler.Filename)))
	return &types.FileUploadResponse{FileID: objectKey, Size: handler.Size}, nil
}
