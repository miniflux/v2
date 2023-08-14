// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pocket // import "miniflux.app/v2/internal/integration/pocket"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/version"
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
	apiEndpoint := "https://getpocket.com/v3/oauth/request"
	requestBody, err := json.Marshal(&createTokenRequest{ConsumerKey: c.consumerKey, RedirectURI: redirectURL})
	if err != nil {
		return "", fmt.Errorf("pocket: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("pocket: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("pocket: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("pocket: unable get request token: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	var result createTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("pocket: unable to decode response: %v", err)
	}

	if result.Code == "" {
		return "", errors.New("pocket: request token is empty")
	}

	return result.Code, nil
}

// AccessToken fetches a new access token once the end-user authorized the application.
func (c *Connector) AccessToken(requestToken string) (string, error) {
	apiEndpoint := "https://getpocket.com/v3/oauth/authorize"
	requestBody, err := json.Marshal(&authorizeRequest{ConsumerKey: c.consumerKey, Code: requestToken})
	if err != nil {
		return "", fmt.Errorf("pocket: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("pocket: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("pocket: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("pocket: unable get access token: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	var result authorizeReponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("pocket: unable to decode response: %v", err)
	}

	if result.AccessToken == "" {
		return "", errors.New("pocket: access token is empty")
	}

	return result.AccessToken, nil
}

// AuthorizationURL returns the authorization URL for the end-user.
func (c *Connector) AuthorizationURL(requestToken, redirectURL string) string {
	return fmt.Sprintf(
		"https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s",
		requestToken,
		redirectURL,
	)
}

type createTokenRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectURI string `json:"redirect_uri"`
}

type createTokenResponse struct {
	Code string `json:"code"`
}

type authorizeRequest struct {
	ConsumerKey string `json:"consumer_key"`
	Code        string `json:"code"`
}

type authorizeReponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}
