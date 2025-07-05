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

	"github.com/pizza-nz/file-uploader/types"
	"github.com/pizza-nz/file-uploader/utils"
)

var (
	_, b, _, _ = runtime.Caller(0)
	RootPath   = filepath.Join(filepath.Dir(b), "../..")
)

type FileUploadService interface {
	CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error)
}

type FileUploadServiceImpl struct {
}

func NewFileUploadService() FileUploadService {
	return &FileUploadServiceImpl{}
}

func (s *FileUploadServiceImpl) CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	defer file.Close()

	tempFolderPath := filepath.Join(RootPath, "tempFiles")
	if _, err := os.Stat(tempFolderPath); os.IsNotExist(err) {
		err = os.MkdirAll(tempFolderPath, os.ModePerm)
		if err != nil {
			slog.Error("Error creating temporary folder", "error", err)
			return nil, types.NewAppError("Internal Server Error", "Error in creating the temporary folder", http.StatusInternalServerError, err)
		}
	}
	slog.Info("Creating temporary folder", "path", tempFolderPath)
	tempFileName := fmt.Sprintf("upload-%s-*%s", utils.FileNameWithoutExtension(handler.Filename), filepath.Ext(handler.Filename))

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	if err != nil {
		slog.Error("Error creating temporary file", "error", err)
		return nil, types.NewAppError("Internal Server Error", "Error in creating the file ", http.StatusInternalServerError, err)
	}

	defer tempFile.Close()

	filebytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading file buffer", "error", err)
		return nil, types.NewAppError("Internal Server Error", "Error in reading the file buffer", http.StatusInternalServerError, err)
	}

	_, err = tempFile.Write(filebytes)
	if err != nil {
		slog.Error("Error writing file to disk", "error", err)
		return nil, types.NewAppError("Internal Server Error", "Error writing file to disk", http.StatusInternalServerError, err)
	}

	slog.Info("File uploaded successfully", "filename", handler.Filename)
	return &types.FileUploadResponse{FileID: filepath.Base(tempFile.Name()), Size: handler.Size}, nil
}
