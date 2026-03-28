// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"
	"errors"
	"log/slog"
)

// Manager manages registered OAuth2 providers.
type Manager struct {
	providers map[string]Provider
}

// FindProvider returns the provider registered under the given name,
// or an error if no such provider exists.
func (m *Manager) FindProvider(name string) (Provider, error) {
	if provider, found := m.providers[name]; found {
		return provider, nil
	}

	return nil, errors.New("oauth2 provider not found")
}

// AddProvider registers a provider under the given name.
func (m *Manager) AddProvider(name string, provider Provider) {
	m.providers[name] = provider
}

// NewManager creates a Manager and registers the specified OAuth2 provider.
// The provider argument must be "oidc" or "google".
func NewManager(ctx context.Context, provider, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint string) *Manager {
	m := &Manager{providers: make(map[string]Provider)}

	switch provider {
	case "oidc":
		if clientSecret == "" {
			slog.Warn("OIDC client secret is empty or missing.")
		}

		if oidcProvider, err := NewOidcProvider(ctx, clientID, clientSecret, redirectURL, oidcDiscoveryEndpoint); err != nil {
			slog.Error("Failed to initialize OIDC provider",
				slog.Any("error", err),
			)
		} else {
			m.AddProvider("oidc", oidcProvider)
		}
	case "google":
		m.AddProvider("google", NewGoogleProvider(clientID, clientSecret, redirectURL))
	default:
		slog.Error("Unsupported OAuth2 provider",
			slog.String("provider", provider),
		)
	}

	return m
}
