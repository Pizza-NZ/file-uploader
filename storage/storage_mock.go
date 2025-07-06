package storage

import (
	"context"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

// MockFileStorage is a mock implementation of the FileStorage interface for testing.
type MockFileStorage struct {
	mock.Mock
}

var _ FileStorage = (*MockFileStorage)(nil)

func (m *MockFileStorage) Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, file, handler)
	return args.String(0), args.Error(1)
}
