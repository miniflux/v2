// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/oauth2"

	"miniflux.app/v2/internal/crypto"
)

// Authorization holds the OAuth2 authorization URL, state parameter, and PKCE code verifier.
type Authorization struct {
	url          string
	state        string
	codeVerifier string
}

// RedirectURL returns the OAuth2 authorization URL to redirect the user to.
func (a *Authorization) RedirectURL() string {
	return a.url
}

// State returns the random state parameter used for CSRF protection.
func (a *Authorization) State() string {
	return a.state
}

// CodeVerifier returns the PKCE code verifier associated with this authorization.
func (a *Authorization) CodeVerifier() string {
	return a.codeVerifier
}

// GenerateAuthorization creates a new Authorization with a random state and PKCE code challenge
// derived from the given OAuth2 configuration.
func GenerateAuthorization(config *oauth2.Config) *Authorization {
	codeVerifier := crypto.GenerateRandomStringHex(32)
	sum := sha256.Sum256([]byte(codeVerifier))

	state := crypto.GenerateRandomStringHex(24)

	authURL := config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", base64.RawURLEncoding.EncodeToString(sum[:])),
	)

	return &Authorization{
		url:          authURL,
		state:        state,
		codeVerifier: codeVerifier,
	}
}
