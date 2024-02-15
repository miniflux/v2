// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkwarden // import "miniflux.app/v2/internal/integration/linkwarden"

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
	baseURL string
	apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("linkwarden: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/v1/links")
	if err != nil {
		return fmt.Errorf(`linkwarden: invalid API endpoint: %v`, err)
	}

	requestBody, err := json.Marshal(&linkwardenBookmark{
		Url:         entryURL,
		Name:        "",
		Description: "",
		Tags:        []string{},
		Collection:  map[string]interface{}{},
	})

	if err != nil {
		return fmt.Errorf("linkwarden: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("linkwarden: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.AddCookie(&http.Cookie{Name: "__Secure-next-auth.session-token", Value: c.apiKey})
	request.AddCookie(&http.Cookie{Name: "next-auth.session-token", Value: c.apiKey})

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("linkwarden: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("linkwarden: unable to create link: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type linkwardenBookmark struct {
	Url         string                 `json:"url"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Collection  map[string]interface{} `json:"collection"`
}
