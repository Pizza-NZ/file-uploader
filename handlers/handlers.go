package handlers

import (
	"log"
	"net/http"

	"github.com/pizza-nz/file-uploader/services"
	"github.com/pizza-nz/file-uploader/types"
	"github.com/pizza-nz/file-uploader/utils"
)

type FileUploadHandler interface {
	CreateFileUpload(w http.ResponseWriter, r *http.Request)

	GetFileUpload(w http.ResponseWriter, r *http.Request)

	DeleteFileUpload(w http.ResponseWriter, r *http.Request)
}

type FileUploadHandlerImpl struct {
	maxFileSize int64
	service     services.FileUploadService
}

func NewFileUploadHandler(maxFileSize int64) FileUploadHandler {
	return &FileUploadHandlerImpl{
		maxFileSize: maxFileSize,
	}
}

func (h *FileUploadHandlerImpl) CreateFileUpload(w http.ResponseWriter, r *http.Request) {
	log.Printf("New Put request\n")
	r.ParseMultipartForm(h.maxFileSize)

	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		utils.HandleError(w, types.NewAppError("Error Reading File", "User file submitted failed to read", http.StatusBadRequest, err))
		return
	}

	err = h.service.CreateFileUpload(file, handler)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusCreated, "File Uploaded")
}
func (h *FileUploadHandlerImpl) GetFileUpload(w http.ResponseWriter, r *http.Request) {

}
func (h *FileUploadHandlerImpl) DeleteFileUpload(w http.ResponseWriter, r *http.Request) {

}
