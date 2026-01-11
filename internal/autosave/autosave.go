// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package autosave

import (
	"sync"
	"time"
)

// Default autosave timing values
const (
	DefaultIdleTimeout = 2 * time.Second  // Time after last edit before autosave
	DefaultMinInterval = 30 * time.Second // Minimum time between autosaves
)

// AutosaveState contains the current buffer state needed for autosave
type AutosaveState struct {
	Lines    [][]rune
	Path     string // empty for unnamed buffers
	EOL      string
	Encoding string
	Modified bool
}

// StateProvider is a function that returns the current buffer state
type StateProvider func() AutosaveState

// ErrorCallback is called when autosave completes (with nil on success)
type ErrorCallback func(err error)

// Manager handles automatic saving of buffer content.
// It uses an idle-based trigger with debouncing to minimize disk I/O.
type Manager struct {
	enabled      bool
	idleTimeout  time.Duration
	minInterval  time.Duration
	lastAutosave time.Time
	idleTimer    *time.Timer
	mu           sync.Mutex

	// Callbacks
	getState   StateProvider
	onAutosave ErrorCallback

	// Track the path we're autosaving for
	currentPath string
}

// NewManager creates a new autosave manager with the given configuration.
func NewManager(enabled bool, idleTimeout, minInterval time.Duration) *Manager {
	if idleTimeout <= 0 {
		idleTimeout = DefaultIdleTimeout
	}
	if minInterval <= 0 {
		minInterval = DefaultMinInterval
	}

	return &Manager{
		enabled:     enabled,
		idleTimeout: idleTimeout,
		minInterval: minInterval,
	}
}

// SetStateProvider sets the callback to get current buffer state
func (m *Manager) SetStateProvider(provider StateProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getState = provider
}

// SetErrorCallback sets the callback for autosave completion
func (m *Manager) SetErrorCallback(callback ErrorCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onAutosave = callback
}

// SetEnabled enables or disables autosave
func (m *Manager) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = enabled

	if !enabled && m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}
}

// IsEnabled returns whether autosave is enabled
func (m *Manager) IsEnabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enabled
}

// NotifyEdit should be called whenever the buffer is modified.
// It resets the idle timer, triggering an autosave after the idle timeout.
func (m *Manager) NotifyEdit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled || m.getState == nil {
		return
	}

	// Reset idle timer
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}

	m.idleTimer = time.AfterFunc(m.idleTimeout, func() {
		m.tryAutosave()
	})
}

// tryAutosave attempts to write the autosave file if conditions are met
func (m *Manager) tryAutosave() {
	m.mu.Lock()

	if !m.enabled || m.getState == nil {
		m.mu.Unlock()
		return
	}

	// Check minimum interval
	if time.Since(m.lastAutosave) < m.minInterval {
		m.mu.Unlock()
		return
	}

	getState := m.getState
	onAutosave := m.onAutosave
	m.mu.Unlock()

	// Get state outside the lock to avoid blocking
	state := getState()

	// Skip if not modified or no path (unnamed buffer)
	if !state.Modified || state.Path == "" {
		return
	}

	// Perform the autosave
	err := WriteAutosave(state.Path, state.Lines, state.EOL, state.Encoding)

	m.mu.Lock()
	if err == nil {
		m.lastAutosave = time.Now()
		m.currentPath = state.Path
	}
	m.mu.Unlock()

	// Notify callback
	if onAutosave != nil {
		onAutosave(err)
	}
}

// ForceAutosave immediately performs an autosave if the buffer is modified.
// This bypasses the idle timer and minimum interval checks.
func (m *Manager) ForceAutosave() error {
	m.mu.Lock()
	if !m.enabled || m.getState == nil {
		m.mu.Unlock()
		return nil
	}

	getState := m.getState
	m.mu.Unlock()

	state := getState()
	if !state.Modified || state.Path == "" {
		return nil
	}

	err := WriteAutosave(state.Path, state.Lines, state.EOL, state.Encoding)

	m.mu.Lock()
	if err == nil {
		m.lastAutosave = time.Now()
		m.currentPath = state.Path
	}
	m.mu.Unlock()

	return err
}

// ClearAutosave removes the autosave file for the given path.
// This should be called after a successful explicit save.
func (m *Manager) ClearAutosave(path string) error {
	if path == "" {
		return nil
	}
	return DeleteAutosave(path)
}

// ClearCurrentAutosave removes the autosave file for the current file being edited.
func (m *Manager) ClearCurrentAutosave() error {
	m.mu.Lock()
	path := m.currentPath
	m.mu.Unlock()

	if path == "" {
		return nil
	}
	return DeleteAutosave(path)
}

// UpdatePath updates the current path when a file is loaded or saved as.
// If there was an autosave for the old path, it is not automatically deleted.
func (m *Manager) UpdatePath(newPath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentPath = newPath
}

// Stop stops the autosave manager and cancels any pending timers.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}
}

// Reset resets the autosave state, clearing the last autosave time.
// This is useful when loading a new file.
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}
	m.lastAutosave = time.Time{}
	m.currentPath = ""
}
