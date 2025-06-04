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

type successResponse struct {
	CreatedAt string `json:"createdAt"`
	Content   struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	}
}

type Client interface {
	SaveUrl(url string) error
}

type client struct {
	wrapped     *http.Client
	apiEndpoint string
	apiToken    string
}

func NewClient(apiToken string, apiEndpoint string) Client {
	return &client{wrapped: &http.Client{Timeout: defaultClientTimeout}, apiEndpoint: apiEndpoint, apiToken: apiToken}
}

func (c *client) SaveUrl(url string) error {
	var payload = map[string]interface{}{
		"type": "link",
		"url":  url,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	resp, err := c.wrapped.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("karakeep: failed to parse response: %s", err)
	}

	if resp.StatusCode >= 400 {
		var errResponse errorResponse
		if err = json.Unmarshal(b, &errResponse); err != nil {
			return fmt.Errorf("karakeep: failed to save URL: status=%d %s", resp.StatusCode, string(b))
		}
		return fmt.Errorf("karakeep: failed to save URL: status=%d errorcode=%s %s", resp.StatusCode, errResponse.Code, errResponse.Error)
	}

	var successReponse successResponse
	if err = json.Unmarshal(b, &successReponse); err != nil {
		return fmt.Errorf("karakeep: failed to parse response, however the request appears successful, is the url correct?: status=%d %s", resp.StatusCode, string(b))
	}

	return nil
}
