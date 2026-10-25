package todo

import (
	"github.com/google/uuid"
	"github.com/sriniously/go-tasker/internal/model"
)

type TodoAttachment struct {
	model.Base
	TodoID      uuid.UUID `json:"todo_id" db:"todo_id"`
	Name        string    `json:"name" db:"name"`
	UploadedBy  string    `json:"uploaded_by" db:"uploaded_by"`
	DownloadKey string    `json:"download_key" db:"download_key"`
	FileSize    *int64    `json:"file_size" db:"file_size"`
	MimeType    *string   `json:"mime_type" db:"mime_type"`
}
