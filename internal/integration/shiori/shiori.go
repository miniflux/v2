// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package shiori // import "miniflux.app/v2/internal/integration/shiori"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL  string
	username string
	password string
}

func NewClient(baseURL, username, password string) *Client {
	return &Client{baseURL: baseURL, username: username, password: password}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.username == "" || c.password == "" {
		return fmt.Errorf("shiori: missing base URL, username or password")
	}

	sessionID, err := c.authenticate()
	if err != nil {
		return fmt.Errorf("shiori: unable to authenticate: %v", err)
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks")
	if err != nil {
		return fmt.Errorf("shiori: invalid API endpoint: %v", err)
	}

	requestBody, err := json.Marshal(&addBookmarkRequest{
		URL:           entryURL,
		Title:         entryTitle,
		CreateArchive: true,
	})

	if err != nil {
		return fmt.Errorf("shiori: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("shiori: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("X-Session-Id", sessionID)

	httpClient := &http.Client{Timeout: defaultClientTimeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("shiori: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("shiori: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

func (c *Client) authenticate() (sessionID string, err error) {
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/login")
	if err != nil {
		return "", fmt.Errorf("shiori: invalid API endpoint: %v", err)
	}

	requestBody, err := json.Marshal(&authRequest{Username: c.username, Password: c.password})
	if err != nil {
		return "", fmt.Errorf("shiori: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("shiori: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("shiori: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("shiori: unable to authenticate: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	var authResponse authResponse
	if err := json.NewDecoder(response.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("shiori: unable to decode response: %v", err)
	}

	return authResponse.SessionID, nil
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	SessionID string `json:"session"`
}

type addBookmarkRequest struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	CreateArchive bool   `json:"createArchive"`
}
