package storage

import (
	"context"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

// MockFileStorage is a mock implementation of the FileStorage interface for local development.
// It simulates file uploads by logging the file details to the console.

type MockFileStorage struct {
	mock.Mock
}

func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{}
}

func (m *MockFileStorage) Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, file, handler)
	return args.String(0), args.Error(1)
}