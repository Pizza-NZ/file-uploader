package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// MockFileStorage is a mock implementation of the FileStorage interface for local development.
// It simulates file uploads by logging the file details to the console.

type MockFileStorage struct {
	mock.Mock
	ShouldFail bool // If true, simulates a failure during upload
}

func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{}
}

func (m *MockFileStorage) Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	_, span := otel.Tracer("file-uploader/storage").Start(ctx, "MockFileStorage.Upload")
	defer span.End()

	if m.ShouldFail {
		err := fmt.Errorf("mock upload failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Mock failure")
		return "", err
	}

	args := m.Called(ctx, file, handler)
	return args.String(0), args.Error(1)
}
