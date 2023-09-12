// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"
	"encoding/json"
	"fmt"

	"miniflux.app/v2/internal/model"

	"golang.org/x/oauth2"
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

func NewGoogleProvider(clientID, clientSecret, redirectURL string) *googleProvider {
	return &googleProvider{clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL}
}

func (g *googleProvider) GetConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  g.redirectURL,
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func (g *googleProvider) GetUserExtraKey() string {
	return "google_id"
}

func (g *googleProvider) GetProfile(ctx context.Context, code, codeVerifier string) (*Profile, error) {
	conf := g.GetConfig()
	token, err := conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("google: failed to exchange token: %w", err)
	}

	client := conf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google: failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var user googleProfile
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("google: unable to unserialize Google profile: %w", err)
	}

	profile := &Profile{Key: g.GetUserExtraKey(), ID: user.Sub, Username: user.Email}
	return profile, nil
}

func (g *googleProvider) PopulateUserCreationWithProfileID(user *model.UserCreationRequest, profile *Profile) {
	user.GoogleID = profile.ID
}

func (g *googleProvider) PopulateUserWithProfileID(user *model.User, profile *Profile) {
	user.GoogleID = profile.ID
}

func (g *googleProvider) UnsetUserProfileID(user *model.User) {
	user.GoogleID = ""
}
