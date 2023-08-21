// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pocket // import "miniflux.app/v2/internal/integration/pocket"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	consumerKey string
	accessToken string
}

func NewClient(consumerKey, accessToken string) *Client {
	return &Client{consumerKey, accessToken}
}

func (c *Client) AddURL(entryURL, entryTitle string) error {
	if c.consumerKey == "" || c.accessToken == "" {
		return fmt.Errorf("pocket: missing consumer key or access token")
	}

	apiEndpoint := "https://getpocket.com/v3/add"
	requestBody, err := json.Marshal(&createItemRequest{
		AccessToken: c.accessToken,
		ConsumerKey: c.consumerKey,
		Title:       entryTitle,
		URL:         entryURL,
	})
	if err != nil {
		return fmt.Errorf("pocket: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("pocket: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("pocket: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("pocket: unable to create item: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type createItemRequest struct {
	AccessToken string `json:"access_token"`
	ConsumerKey string `json:"consumer_key"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url"`
}
