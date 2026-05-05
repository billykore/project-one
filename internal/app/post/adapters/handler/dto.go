package handler

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required,min=10"`
	Tags    []string `json:"tags"`
}

type CreatePostResponse struct {
	Message     string `json:"message"`
	RedirectURL string `json:"redirect_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
