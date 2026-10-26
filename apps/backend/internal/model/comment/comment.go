package comment

import (
	"github.com/google/uuid"
	"github.com/vinaykvv00/go-tasker/internal/model"
)

type Comment struct {
	model.Base
	TodoID  uuid.UUID `json:"todoId" db:"todo_id"`
	UserID  string    `json:"userId" db:"user_id"`
	Content string    `json:"content" db:"content"`
}
