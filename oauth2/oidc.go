// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package oauth2 // import "miniflux.app/oauth2"

import (
	"context"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type oidcProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	provider     *oidc.Provider
}

func (o oidcProvider) GetUserExtraKey() string {
	return "oidc_id" // FIXME? add extra options key to allow multiple OIDC providers each with their own extra key?
}

func (o oidcProvider) GetRedirectURL(state string) string {
	return o.config().AuthCodeURL(state)
}

func (o oidcProvider) GetProfile(ctx context.Context, code string) (*Profile, error) {
	conf := o.config()
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	userInfo, err := o.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, err
	}

	profile := &Profile{Key: o.GetUserExtraKey(), ID: userInfo.Subject, Username: userInfo.Email}
	return profile, nil
}

func (o oidcProvider) config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  o.redirectURL,
		ClientID:     o.clientID,
		ClientSecret: o.clientSecret,
		Scopes:       []string{"openid", "email"},
		Endpoint:     o.provider.Endpoint(),
	}
}

func newOidcProvider(ctx context.Context, clientID, clientSecret, redirectURL, discoveryEndpoint string) (*oidcProvider, error) {
	provider, err := oidc.NewProvider(ctx, discoveryEndpoint)
	if err != nil {
		return nil, err
	}

	return &oidcProvider{clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL, provider: provider}, nil
}
