package services

import (
	"fmt"
	"io"
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
	tempFileName := fmt.Sprintf("upload-%s-*%s", utils.FileNameWithoutExtension(handler.Filename), filepath.Ext(handler.Filename))

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	if err != nil {
		return types.NewAppError("Internal Server Error", "Error in creating the file ", http.StatusInternalServerError, err)
	}

	defer tempFile.Close()

	filebytes, err := io.ReadAll(file)
	if err != nil {
		return types.NewAppError("Internal Server Error", "Error in reading the file buffer", http.StatusInternalServerError, err)
	}

	tempFile.Write(filebytes)
	return nil
}
