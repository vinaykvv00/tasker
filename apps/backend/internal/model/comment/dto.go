package comment

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Data Transfer Objects (DTOs) for comment-related operations

// Note: The naming convention for the payloads has been standardized to use "ID" consistently.

// payload for adding a new comment to a todo item
type AddCommentPayload struct {
	TodoID  uuid.UUID `param:"id" validate:"required,uuid"`
	Content string    `json:"content" validate:"required,min=1,max=1000"`
}

func (p *AddCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// payload for fetching all comments associated with a specific todo item
type GetCommentsByTodoIDPayload struct {
	TodoID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *GetCommentsByTodoIDPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// updating an existing comment
type UpdateCommentPayload struct {
	ID      uuid.UUID `param:"id" validate:"required,uuid"`
	Content string    `json:"content" validate:"required,min=1,max=1000"`
}

func (p *UpdateCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// payload for deleting an existing comment
type DeleteCommentPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
