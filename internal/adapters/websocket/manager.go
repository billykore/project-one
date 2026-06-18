package websocket

import (
	"errors"
	"sync"

	"github.com/billykore/project-one/internal/api/dto"
	gws "github.com/gorilla/websocket"
)

var ErrUserNotConnected = errors.New("user is not connected")

type managedConn struct {
	conn *gws.Conn
	mu   sync.Mutex
}

type Manager struct {
	mu          sync.RWMutex
	connections map[int]*managedConn
}

func NewManager() *Manager {
	return &Manager{connections: make(map[int]*managedConn)}
}

func (m *Manager) Register(userID int, conn *gws.Conn) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if prev, ok := m.connections[userID]; ok && prev != nil && prev.conn != nil {
		_ = prev.conn.Close()
	}

	m.connections[userID] = &managedConn{conn: conn}
	return nil
}

func (m *Manager) Unregister(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, userID)
}

func (m *Manager) Send(notification *dto.NotificationResponse) error {
	m.mu.RLock()
	mc, ok := m.connections[notification.UserID]
	m.mu.RUnlock()
	if !ok || mc == nil || mc.conn == nil {
		return ErrUserNotConnected
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.conn.WriteJSON(notification)
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for userID, mc := range m.connections {
		if mc != nil && mc.conn != nil {
			_ = mc.conn.Close()
		}
		delete(m.connections, userID)
	}

	return nil
}

func (m *Manager) ConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}
