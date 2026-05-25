package session

import (
	"sync"

	"media-jukebox-backend/internal/model"
)

type Manager struct {
	mu      sync.Mutex
	session model.Session
}

func New() *Manager {
	return &Manager{
		session: model.Session{
			State: "idle",
		},
	}
}

func (m *Manager) SetPlaying(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.session.CurrentID = id
	m.session.State = "playing"
}

func (m *Manager) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.session.State = "paused"
}

func (m *Manager) Idle() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.session.CurrentID = ""
	m.session.State = "idle"
}

func (m *Manager) Get() model.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.session
}
