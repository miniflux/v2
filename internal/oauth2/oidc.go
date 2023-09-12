// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"
	"errors"
	"fmt"

	"miniflux.app/v2/internal/model"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrEmptyUsername = errors.New("oidc: username is empty")
)

type oidcProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	provider     *oidc.Provider
}

func NewOidcProvider(ctx context.Context, clientID, clientSecret, redirectURL, discoveryEndpoint string) (*oidcProvider, error) {
	provider, err := oidc.NewProvider(ctx, discoveryEndpoint)
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to initialize provider %q: %w`, discoveryEndpoint, err)
	}

	return &oidcProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		provider:     provider,
	}, nil
}

func (o *oidcProvider) GetUserExtraKey() string {
	return "openid_connect_id"
}

func (o *oidcProvider) GetConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  o.redirectURL,
		ClientID:     o.clientID,
		ClientSecret: o.clientSecret,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     o.provider.Endpoint(),
	}
}

func (o *oidcProvider) GetProfile(ctx context.Context, code, codeVerifier string) (*Profile, error) {
	conf := o.GetConfig()
	token, err := conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to exchange token: %w`, err)
	}

	userInfo, err := o.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to get user info: %w`, err)
	}

	profile := &Profile{
		Key: o.GetUserExtraKey(),
		ID:  userInfo.Subject,
	}

	var userClaims userClaims
	if err := userInfo.Claims(&userClaims); err != nil {
		return nil, fmt.Errorf(`oidc: failed to parse user claims: %w`, err)
	}

	for _, value := range []string{userClaims.Email, userClaims.PreferredUsername, userClaims.Name, userClaims.Profile} {
		if value != "" {
			profile.Username = value
			break
		}
	}

	if profile.Username == "" {
		return nil, ErrEmptyUsername
	}

	return profile, nil
}

func (o *oidcProvider) PopulateUserCreationWithProfileID(user *model.UserCreationRequest, profile *Profile) {
	user.OpenIDConnectID = profile.ID
}

func (o *oidcProvider) PopulateUserWithProfileID(user *model.User, profile *Profile) {
	user.OpenIDConnectID = profile.ID
}

func (o *oidcProvider) UnsetUserProfileID(user *model.User) {
	user.OpenIDConnectID = ""
}

type userClaims struct {
	Email             string `json:"email"`
	Profile           string `json:"profile"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}
