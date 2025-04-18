package groups

import (
	"errors"
	"sync"
)

// Manager manages chat groups.
type Manager struct {
	mutex  sync.RWMutex
	groups map[string]map[string]bool // groupName → set of userIDs
}

// NewManager constructs a group manager.
func NewManager() *Manager {
	return &Manager{groups: make(map[string]map[string]bool)}
}

// CreateGroup creates a new group.
func (m *Manager) CreateGroup(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.groups[name]; ok {
		return errors.New("group already exists")
	}
	m.groups[name] = make(map[string]bool)
	return nil
}

// AddMember adds a user to a group.
func (m *Manager) AddMember(name, userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	set, ok := m.groups[name]
	if !ok {
		return errors.New("group not found")
	}
	set[userID] = true
	return nil
}

// RemoveMember removes a user from a group.
func (m *Manager) RemoveMember(name, userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	set, ok := m.groups[name]
	if !ok {
		return errors.New("group not found")
	}
	delete(set, userID)
	return nil
}

// Members returns all userIDs in a group.
func (m *Manager) Members(name string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	set, ok := m.groups[name]
	if !ok {
		return nil, errors.New("group not found")
	}
	var ids []string
	for uid := range set {
		ids = append(ids, uid)
	}
	return ids, nil
}
