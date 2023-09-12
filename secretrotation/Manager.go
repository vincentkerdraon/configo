package secretrotation

import (
	"sync"
)

type (
	Manager struct {
		rs RotatingSecret
		sync.RWMutex
	}
)

func New() *Manager {
	return &Manager{
		RWMutex: sync.RWMutex{},
	}
}

func (m *Manager) Set(rs RotatingSecret) error {
	if err := rs.Validate(); err != nil {
		return InvalidSecretError{Err: err}
	}

	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()

	m.rs = rs
	return nil
}

func (m *Manager) RotatingSecret() (RotatingSecret, error) {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()

	//Safer to force checking the error type than not returning an error and risking to forget to Validate().
	if err := m.rs.Validate(); err != nil {
		return RotatingSecret{}, MissingInitValuesError{}
	}

	return m.rs, nil
}

// Current returns the secret to use when the consumer is calling the provider.
func (m *Manager) Current() (Secret, error) {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()

	//Safer to force checking the error type than not returning an error and risking to forget to Validate().
	if err := m.rs.Validate(); err != nil {
		return "", MissingInitValuesError{}
	}

	return m.rs.Current, nil
}

// Allowed checks if a given key match the secrets.
func (m *Manager) Allowed(in Secret) bool {
	var rs RotatingSecret
	func() {
		m.RWMutex.RLock()
		defer m.RWMutex.RUnlock()
		rs = m.rs
	}()

	return rs.Allowed(in)
}

// AllowedNonConstant checks if a given key match the secrets.
// This is NOT using the crypto security on timing attacks.
// This is faster than Allowed()
func (m *Manager) AllowedNonConstant(in Secret) bool {
	var rs RotatingSecret
	func() {
		m.RWMutex.RLock()
		defer m.RWMutex.RUnlock()
		rs = m.rs
	}()

	return rs.AllowedNonConstant(in)
}
