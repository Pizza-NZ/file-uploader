package storage

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/pizza-nz/file-uploader/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// S3Storage implements the FileStorage interface for AWS S3.
type S3Storage struct {
	client     *s3.Client
	bucketName string
}

var _ FileStorage = (*S3Storage)(nil)

// NewS3Storage creates a new S3Storage instance.
func NewS3Storage(ctx context.Context, cfg config.AWSConfig) (FileStorage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		slog.Error("failed to load AWS config", "error", err)
		return nil, err
	}

	return &S3Storage{
		client:     s3.NewFromConfig(awsCfg),
		bucketName: cfg.S3.BucketName,
	}, nil
}

// Upload uploads a file to S3 and returns the object key.
func (s *S3Storage) Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	ctx, span := otel.Tracer("file-uploader/storage").Start(ctx, "Upload")
	defer span.End()
	span.SetAttributes(
		attribute.String("aws.bucket", s.bucketName),
		attribute.String("file.name", handler.Filename),
		attribute.String("file.type", handler.Header.Get("Content-Type")),
	)
	fileID := uuid.New().String()
	fileExtension := filepath.Ext(handler.Filename)
	s3ObjectKey := fmt.Sprintf("%s%s", fileID, fileExtension)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(s3ObjectKey),
		Body:        file,
		ContentType: aws.String(handler.Header.Get("Content-Type")),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "File upload failed")
		slog.Error("Error uploading file to S3", "error", err)
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return s3ObjectKey, nil
}
