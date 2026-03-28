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

// ErrEmptyUsername is returned when the OIDC user profile has no username.
var ErrEmptyUsername = errors.New("oidc: username is empty")

type userClaims struct {
	Email             string `json:"email"`
	Profile           string `json:"profile"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

type oidcProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	provider     *oidc.Provider
}

// NewOidcProvider returns a Provider that authenticates users via OpenID Connect.
// It discovers the OIDC endpoints from the given discovery URL.
func NewOidcProvider(ctx context.Context, clientID, clientSecret, redirectURL, discoveryEndpoint string) (Provider, error) {
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

func (o *oidcProvider) UserExtraKey() string {
	return "openid_connect_id"
}

func (o *oidcProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  o.redirectURL,
		ClientID:     o.clientID,
		ClientSecret: o.clientSecret,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     o.provider.Endpoint(),
	}
}

func (o *oidcProvider) Profile(ctx context.Context, code, codeVerifier string) (*UserProfile, error) {
	conf := o.Config()
	token, err := conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to exchange token: %w`, err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New(`oidc: no id_token in token response`)
	}

	verifier := o.provider.Verifier(&oidc.Config{ClientID: o.clientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to verify id token: %w`, err)
	}

	userInfo, err := o.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf(`oidc: failed to get user info: %w`, err)
	}

	if idToken.Subject != userInfo.Subject {
		return nil, fmt.Errorf(`oidc: id token subject %q does not match userinfo subject %q`, idToken.Subject, userInfo.Subject)
	}

	profile := &UserProfile{
		Key: o.UserExtraKey(),
		ID:  userInfo.Subject,
	}

	var userClaims userClaims
	if err := userInfo.Claims(&userClaims); err != nil {
		return nil, fmt.Errorf(`oidc: failed to parse user claims: %w`, err)
	}

	// Use the first non-empty value from the claims to set the username.
	// The order of preference is: preferred_username, email, name, profile.
	for _, value := range []string{userClaims.PreferredUsername, userClaims.Email, userClaims.Name, userClaims.Profile} {
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

func (o *oidcProvider) PopulateUserCreationWithProfileID(user *model.UserCreationRequest, profile *UserProfile) {
	user.OpenIDConnectID = profile.ID
}

func (o *oidcProvider) PopulateUserWithProfileID(user *model.User, profile *UserProfile) {
	user.OpenIDConnectID = profile.ID
}

func (o *oidcProvider) UserProfileID(user *model.User) string {
	return user.OpenIDConnectID
}

func (o *oidcProvider) UnsetUserProfileID(user *model.User) {
	user.OpenIDConnectID = ""
}
