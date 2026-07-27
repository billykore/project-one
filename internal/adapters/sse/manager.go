package sse

import (
	"errors"
	"sync"

	"github.com/billykore/project-one/internal/api/dto"
)

var ErrUserNotConnected = errors.New("user is not connected")

type Manager struct {
	mu      sync.RWMutex
	clients map[int]map[chan *dto.NotificationResponse]struct{}
	closed  bool
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[int]map[chan *dto.NotificationResponse]struct{}),
	}
}

func (m *Manager) Register(userID int) chan *dto.NotificationResponse {
	msgCh := make(chan *dto.NotificationResponse, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		close(msgCh)
		return msgCh
	}

	if m.clients[userID] == nil {
		m.clients[userID] = make(map[chan *dto.NotificationResponse]struct{})
	}
	m.clients[userID][msgCh] = struct{}{}
	return msgCh
}

func (m *Manager) Unregister(userID int, msgCh chan *dto.NotificationResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.clients[userID]
	if clients == nil {
		return
	}
	if _, ok := clients[msgCh]; !ok {
		return
	}

	delete(clients, msgCh)
	close(msgCh)
	if len(clients) == 0 {
		delete(m.clients, userID)
	}
}

func (m *Manager) Send(notification *dto.NotificationResponse) error {
	m.mu.RLock()
	clients := m.clients[notification.UserID]
	if len(clients) == 0 {
		m.mu.RUnlock()
		return ErrUserNotConnected
	}

	for msgCh := range clients {
		select {
		case msgCh <- notification:
		default:
		}
	}
	m.mu.RUnlock()
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	for userID, clients := range m.clients {
		for msgCh := range clients {
			close(msgCh)
			delete(clients, msgCh)
		}
		delete(m.clients, userID)
	}

	return nil
}

func (m *Manager) ConnectionCount(userID int) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients[userID])
}
