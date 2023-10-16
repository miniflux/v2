// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"
	"errors"
	"log/slog"
)

type Manager struct {
	providers map[string]Provider
}

func (m *Manager) FindProvider(name string) (Provider, error) {
	if provider, found := m.providers[name]; found {
		return provider, nil
	}

	return nil, errors.New("oauth2 provider not found")
}

func (m *Manager) AddProvider(name string, provider Provider) {
	m.providers[name] = provider
}

func NewManager(ctx context.Context, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint string) *Manager {
	m := &Manager{providers: make(map[string]Provider)}
	m.AddProvider("google", NewGoogleProvider(clientID, clientSecret, redirectURL))

	if oidcDiscoveryEndpoint != "" {
		if genericOidcProvider, err := NewOidcProvider(ctx, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint); err != nil {
			slog.Error("Failed to initialize OIDC provider",
				slog.Any("error", err),
			)
		} else {
			m.AddProvider("oidc", genericOidcProvider)
		}
	}

	return m
}
