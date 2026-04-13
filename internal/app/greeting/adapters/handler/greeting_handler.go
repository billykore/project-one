package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/billykore/project-one/internal/app/greeting/adapters/dto"
	"github.com/billykore/project-one/internal/app/greeting/core/ports"
)

type GreetingHandler struct {
	log     *slog.Logger
	service ports.GreetingService
}

// NewGreetingHandler creates a new instance of GreetingHandler.
func NewGreetingHandler(log *slog.Logger, service ports.GreetingService) *GreetingHandler {
	return &GreetingHandler{
		log:     log.With("layer", "handler", "handler", "greeting"),
		service: service,
	}
}

// GetGreeting handles GET /greeting requests.
func (h *GreetingHandler) GetGreeting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	greeting, err := h.service.GetGreeting(r.Context())
	if err != nil {
		h.log.Error("failed to get greeting", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := dto.GreetingResponse{
		Message: greeting.Message,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", "error", err)
	}
}
