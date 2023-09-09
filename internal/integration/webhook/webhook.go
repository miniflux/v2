// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webhook // import "miniflux.app/v2/internal/integration/webhook"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	webhookURL    string
	webhookSecret string
}

func NewClient(webhookURL, webhookSecret string) *Client {
	return &Client{webhookURL, webhookSecret}
}

func (c *Client) SendWebhook(entries model.Entries) error {
	if c.webhookURL == "" {
		return fmt.Errorf(`webhook: missing webhook URL`)
	}

	if len(entries) == 0 {
		return nil
	}

	requestBody, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("webhook: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("webhook: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("X-Miniflux-Signature", crypto.GenerateSHA256Hmac(c.webhookSecret, requestBody))

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("webhook: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("webhook: incorrect response status code: url=%s status=%d", c.webhookURL, response.StatusCode)
	}

	return nil
}
