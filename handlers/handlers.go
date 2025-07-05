package handlers

import (
	"log/slog"
	"net/http"

	"github.com/pizza-nz/file-uploader/services"
	"github.com/pizza-nz/file-uploader/types"
	"github.com/pizza-nz/file-uploader/utils"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, r, http.StatusOK, "OK")
}

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
		service:     services.NewFileUploadService(),
	}
}

func (h *FileUploadHandlerImpl) CreateFileUpload(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		panic("FileUploadService is not initialized")
	}
	slog.Info("New Put request", "requestID", r.Header.Get("X-Request-ID"))
	r.ParseMultipartForm(h.maxFileSize)

	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		utils.HandleError(w, r, types.NewAppError("Error Reading File", "User file submitted failed to read", http.StatusBadRequest, err))
		return
	}

	err = h.service.CreateFileUpload(file, handler)
	if err != nil {
		utils.HandleError(w, r, err)
		return
	}

	utils.JSONResponse(w, r, http.StatusCreated, "File Uploaded")
}
func (h *FileUploadHandlerImpl) GetFileUpload(w http.ResponseWriter, r *http.Request) {

}
func (h *FileUploadHandlerImpl) DeleteFileUpload(w http.ResponseWriter, r *http.Request) {

}
