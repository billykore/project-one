package dto

type CreateCommentRequest struct {
	ID      int    `param:"id" validate:"required,min=1"`
	Content string `json:"content" validate:"required,min=1"`
}
