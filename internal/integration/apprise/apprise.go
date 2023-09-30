// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apprise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	servicesURL string
	baseURL     string
}

func NewClient(serviceURL, baseURL string) *Client {
	return &Client{serviceURL, baseURL}
}

func (c *Client) SendNotification(entry *model.Entry) error {
	if c.baseURL == "" || c.servicesURL == "" {
		return fmt.Errorf("apprise: missing base URL or services URL")
	}

	message := "[" + entry.Title + "]" + "(" + entry.URL + ")" + "\n\n"
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/notify")
	if err != nil {
		return fmt.Errorf(`apprise: invalid API endpoint: %v`, err)
	}

	requestBody, err := json.Marshal(map[string]any{
		"urls": c.servicesURL,
		"body": message,
	})
	if err != nil {
		return fmt.Errorf("apprise: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("apprise: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("apprise: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("apprise: unable to send a notification: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}
