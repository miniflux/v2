// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/model"

	"golang.org/x/oauth2"
)

// Google OAuth2 API documentation: https://developers.google.com/identity/protocols/oauth2
const (
	googleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)

type googleProfile struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

type googleProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// NewGoogleProvider returns a Provider that authenticates users via Google OAuth2.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) Provider {
	return &googleProvider{clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL}
}

func (g *googleProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  g.redirectURL,
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  googleAuthURL,
			TokenURL: googleTokenURL,
		},
	}
}

func (g *googleProvider) UserExtraKey() string {
	return "google_id"
}

func (g *googleProvider) Profile(ctx context.Context, code, codeVerifier string) (*UserProfile, error) {
	conf := g.Config()
	token, err := conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("google: failed to exchange token: %w", err)
	}

	client := conf.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("google: failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google: unexpected status code %d from userinfo endpoint", resp.StatusCode)
	}

	var user googleProfile
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("google: unable to unserialize Google profile: %w", err)
	}

	return &UserProfile{Key: g.UserExtraKey(), ID: user.Sub, Username: user.Email}, nil
}

func (g *googleProvider) PopulateUserCreationWithProfileID(user *model.UserCreationRequest, profile *UserProfile) {
	user.GoogleID = profile.ID
}

func (g *googleProvider) PopulateUserWithProfileID(user *model.User, profile *UserProfile) {
	user.GoogleID = profile.ID
}

func (g *googleProvider) UserProfileID(user *model.User) string {
	return user.GoogleID
}

func (g *googleProvider) UnsetUserProfileID(user *model.User) {
	user.GoogleID = ""
}
