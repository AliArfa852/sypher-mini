package secrets

import (
	"os"
)

// Backend is the secrets backend type.
type Backend string

const (
	BackendEnv      Backend = "env"
	BackendKeychain Backend = "keychain"
)

// Resolver resolves secrets from env or keychain.
type Resolver struct {
	backend Backend
}

// NewResolver creates a secrets resolver.
func NewResolver(backend Backend) *Resolver {
	if backend != BackendKeychain {
		backend = BackendEnv
	}
	return &Resolver{backend: backend}
}

// Get returns the secret for key. For "keychain" backend, falls back to env
// (full keychain impl would use platform APIs: Windows Credential Manager,
// macOS Keychain, Linux secret-service).
func (r *Resolver) Get(key string) string {
	if r.backend == BackendKeychain {
		// TODO: platform-specific keychain lookup
		// For now, fall back to env
	}
	return os.Getenv(key)
}
