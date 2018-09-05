// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package oauth2 // import "miniflux.app/oauth2"

import "errors"

// Manager handles OAuth2 providers.
type Manager struct {
	providers map[string]Provider
}

// Provider returns the given provider.
func (m *Manager) Provider(name string) (Provider, error) {
	if provider, found := m.providers[name]; found {
		return provider, nil
	}

	return nil, errors.New("oauth2 provider not found")
}

// AddProvider add a new OAuth2 provider.
func (m *Manager) AddProvider(name string, provider Provider) {
	m.providers[name] = provider
}

// NewManager returns a new Manager.
func NewManager(clientID, clientSecret, redirectURL string) *Manager {
	m := &Manager{providers: make(map[string]Provider)}
	m.AddProvider("google", newGoogleProvider(clientID, clientSecret, redirectURL))
	return m
}
