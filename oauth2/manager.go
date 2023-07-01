// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/oauth2"

import (
	"context"
	"errors"

	"miniflux.app/logger"
)

// Manager handles OAuth2 providers.
type Manager struct {
	providers map[string]Provider
}

// FindProvider returns the given provider.
func (m *Manager) FindProvider(name string) (Provider, error) {
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
func NewManager(ctx context.Context, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint string) *Manager {
	m := &Manager{providers: make(map[string]Provider)}
	m.AddProvider("google", newGoogleProvider(clientID, clientSecret, redirectURL))

	if oidcDiscoveryEndpoint != "" {
		if genericOidcProvider, err := newOidcProvider(ctx, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint); err != nil {
			logger.Error("[OAuth2] failed to initialize OIDC provider: %v", err)
		} else {
			m.AddProvider("oidc", genericOidcProvider)
		}
	}

	return m
}
