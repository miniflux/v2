// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package karakeep // import "miniflux.app/v2/internal/integration/karakeep"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type errorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

type saveURLPayload struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Client struct {
	wrapped     *http.Client
	apiEndpoint string
	apiToken    string
}

func NewClient(apiToken string, apiEndpoint string) *Client {
	return &Client{wrapped: &http.Client{Timeout: defaultClientTimeout}, apiEndpoint: apiEndpoint, apiToken: apiToken}
}

func (c *Client) SaveURL(entryURL string) error {
	requestBody, err := json.Marshal(&saveURLPayload{
		Type: "link",
		URL:  entryURL,
	})
	if err != nil {
		return fmt.Errorf("karakeep: unable to encode request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("karakeep: unable to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	resp, err := c.wrapped.Do(req)
	if err != nil {
		return fmt.Errorf("karakeep: unable to send request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("karakeep: failed to parse response: %s", err)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("karakeep: unexpected content type response: %s", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != http.StatusCreated {
		var errResponse errorResponse
		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			return fmt.Errorf("karakeep: unable to parse error response: status=%d body=%s", resp.StatusCode, string(responseBody))
		}
		return fmt.Errorf("karakeep: failed to save URL: status=%d errorcode=%s %s", resp.StatusCode, errResponse.Code, errResponse.Error)
	}

	return nil
}
