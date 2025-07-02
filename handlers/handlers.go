package handlers

import (
	"log"
	"net/http"
)

type FileUploadHandler interface {
	CreateFileUpload(w http.ResponseWriter, r *http.Request)

	GetFileUpload(w http.ResponseWriter, r *http.Request)

	DeleteFileUpload(w http.ResponseWriter, r *http.Request)
}

type FileUploadHandlerImpl struct {
}

func NewFileUploadHandler() FileUploadHandler {
	return &FileUploadHandlerImpl{}
}

func (h *FileUploadHandlerImpl) CreateFileUpload(w http.ResponseWriter, r *http.Request) {
	log.Printf("New Put request")
}
func (h *FileUploadHandlerImpl) GetFileUpload(w http.ResponseWriter, r *http.Request) {

}
func (h *FileUploadHandlerImpl) DeleteFileUpload(w http.ResponseWriter, r *http.Request) {

}
