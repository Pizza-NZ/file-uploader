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
	CreateFileUpload(file multipart.File, handler *multipart.FileHeader) error
}

type FileUploadServiceImpl struct {
}

func (s *FileUploadServiceImpl) CreateFileUpload(file multipart.File, handler *multipart.FileHeader) error {
	defer file.Close()

	tempFolderPath := fmt.Sprintf("%s%s", RootPath, "/tempFiles")
	slog.Info("Creating temporary folder", "path", tempFolderPath)
	tempFileName := fmt.Sprintf("upload-%s-*%s", utils.FileNameWithoutExtension(handler.Filename), filepath.Ext(handler.Filename))

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	if err != nil {
		slog.Error("Error creating temporary file", "error", err)
		return types.NewAppError("Internal Server Error", "Error in creating the file ", http.StatusInternalServerError, err)
	}

	defer tempFile.Close()

	filebytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading file buffer", "error", err)
		return types.NewAppError("Internal Server Error", "Error in reading the file buffer", http.StatusInternalServerError, err)
	}

	tempFile.Write(filebytes)
	slog.Info("File uploaded successfully", "filename", handler.Filename)
	return nil
}