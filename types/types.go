package types

type FileUploadResponse struct {
	FileID string `json:"fileId"`
	Size   int64  `json:"size"`
}
