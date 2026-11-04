package sse

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/notifications/api/dto"
	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/notifications/ports"
)

// Notifier adapts Manager to the notifications RealtimeNotifier port.
type Notifier struct {
	manager *Manager
}

var _ ports.RealtimeNotifier = (*Notifier)(nil)

// NewNotifier creates a RealtimeNotifier backed by manager.
func NewNotifier(manager *Manager) *Notifier {
	return &Notifier{manager: manager}
}

// Notify sends the notification to every live connection of its recipient.
func (n *Notifier) Notify(_ context.Context, notification domain.Notification) (bool, error) {
	err := n.manager.Send(dto.NotificationResponseFromDomain(notification))
	if errors.Is(err, ErrUserNotConnected) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
