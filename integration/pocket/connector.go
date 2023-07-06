// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pocket // import "miniflux.app/integration/pocket"

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	"miniflux.app/http/client"
)

// Connector manages the authorization flow with Pocket to get a personal access token.
type Connector struct {
	consumerKey string
}

// NewConnector returns a new Pocket Connector.
func NewConnector(consumerKey string) *Connector {
	return &Connector{consumerKey}
}

// RequestToken fetches a new request token from Pocket API.
func (c *Connector) RequestToken(redirectURL string) (string, error) {
	type req struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}

	clt := client.New("https://getpocket.com/v3/oauth/request")
	response, err := clt.PostJSON(&req{ConsumerKey: c.consumerKey, RedirectURI: redirectURL})
	if err != nil {
		return "", fmt.Errorf("pocket: unable to fetch request token: %v", err)
	}

	if response.HasServerFailure() {
		return "", fmt.Errorf("pocket: unable to fetch request token, status=%d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("pocket: unable to read response body: %v", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("pocket: unable to parse response: %v", err)
	}

	code := values.Get("code")
	if code == "" {
		return "", errors.New("pocket: code is empty")
	}

	return code, nil
}

// AccessToken fetches a new access token once the end-user authorized the application.
func (c *Connector) AccessToken(requestToken string) (string, error) {
	type req struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	clt := client.New("https://getpocket.com/v3/oauth/authorize")
	response, err := clt.PostJSON(&req{ConsumerKey: c.consumerKey, Code: requestToken})
	if err != nil {
		return "", fmt.Errorf("pocket: unable to fetch access token: %v", err)
	}

	if response.HasServerFailure() {
		return "", fmt.Errorf("pocket: unable to fetch access token, status=%d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("pocket: unable to read response body: %v", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("pocket: unable to parse response: %v", err)
	}

	token := values.Get("access_token")
	if token == "" {
		return "", errors.New("pocket: access_token is empty")
	}

	return token, nil
}

// AuthorizationURL returns the authorization URL for the end-user.
func (c *Connector) AuthorizationURL(requestToken, redirectURL string) string {
	return fmt.Sprintf(
		"https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s",
		requestToken,
		redirectURL,
	)
}
