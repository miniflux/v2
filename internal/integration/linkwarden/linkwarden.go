// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkwarden // import "miniflux.app/v2/internal/integration/linkwarden"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL      string
	apiKey       string
	collectionID *int64
}

type linkwardenCollection struct {
	ID *int64 `json:"id"`
}

type linkwardenRequest struct {
	URL        string                `json:"url"`
	Name       string                `json:"name"`
	Collection *linkwardenCollection `json:"collection,omitempty"`
}

func NewClient(baseURL, apiKey string, collectionID *int64) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, collectionID: collectionID}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return errors.New("linkwarden: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/v1/links")
	if err != nil {
		return fmt.Errorf(`linkwarden: invalid API endpoint: %v`, err)
	}

	payload := linkwardenRequest{
		URL:  entryURL,
		Name: entryTitle,
	}

	if c.collectionID != nil {
		payload.Collection = &linkwardenCollection{ID: c.collectionID}
	}

	requestBody, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("linkwarden: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("linkwarden: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("linkwarden: unable to send request: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("linkwarden: unable to read response body: %v", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("linkwarden: unable to create link: status=%d body=%s", response.StatusCode, string(responseBody))
	}

	return nil
}
