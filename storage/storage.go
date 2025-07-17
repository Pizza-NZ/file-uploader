package storage

import (
	"context"
	"mime/multipart"
)

// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error)
}
