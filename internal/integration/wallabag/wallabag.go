// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package wallabag // import "miniflux.app/v2/internal/integration/wallabag"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
	onlyURL      bool
}

func NewClient(baseURL, clientID, clientSecret, username, password string, onlyURL bool) *Client {
	return &Client{baseURL, clientID, clientSecret, username, password, onlyURL}
}

func (c *Client) CreateEntry(entryURL, entryTitle, entryContent string) error {
	if c.baseURL == "" || c.clientID == "" || c.clientSecret == "" || c.username == "" || c.password == "" {
		return fmt.Errorf("wallabag: missing base URL, client ID, client secret, username or password")
	}

	accessToken, err := c.getAccessToken()
	if err != nil {
		return err
	}

	return c.createEntry(accessToken, entryURL, entryTitle, entryContent)
}

func (c *Client) createEntry(accessToken, entryURL, entryTitle, entryContent string) error {
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/entries.json")
	if err != nil {
		return fmt.Errorf("wallbag: unable to generate entries endpoint: %v", err)
	}

	if c.onlyURL {
		entryContent = ""
	}

	requestBody, err := json.Marshal(&createEntryRequest{
		URL:     entryURL,
		Title:   entryTitle,
		Content: entryContent,
	})
	if err != nil {
		return fmt.Errorf("wallbag: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("wallbag: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("wallabag: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("wallabag: unable to get access token: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

func (c *Client) getAccessToken() (string, error) {
	values := url.Values{}
	values.Add("grant_type", "password")
	values.Add("client_id", c.clientID)
	values.Add("client_secret", c.clientSecret)
	values.Add("username", c.username)
	values.Add("password", c.password)

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/oauth/v2/token")
	if err != nil {
		return "", fmt.Errorf("wallbag: unable to generate token endpoint: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("wallbag: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("wallabag: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("wallabag: unable to get access token: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	var responseBody tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("wallabag: unable to decode token response: %v", err)
	}

	return responseBody.AccessToken, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	Expires      int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type createEntryRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}
