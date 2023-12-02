// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package omnivore // import "miniflux.app/v2/internal/integration/omnivore"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

const defaultApiEndpoint = "https://api-prod.omnivore.app/api/graphql"

var mutation = `
mutation SaveUrl($input: SaveUrlInput!) {
  saveUrl(input: $input) {
    ... on SaveSuccess {
      url
      clientRequestId
    }
    ... on SaveError {
      errorCodes
      message
    }
  }
}
`

type SaveUrlInput struct {
	ClientRequestId string `json:"clientRequestId"`
	Source          string `json:"source"`
	Url             string `json:"url"`
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
	if apiEndpoint == "" {
		apiEndpoint = defaultApiEndpoint
	}

	return &client{wrapped: &http.Client{Timeout: defaultClientTimeout}, apiEndpoint: apiEndpoint, apiToken: apiToken}
}

func (c *client) SaveUrl(url string) error {
	var payload = map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"clientRequestId": uuid.New().String(),
				"source":          "api",
				"url":             url,
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	resp, err := c.wrapped.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("omnivore: failed to save URL: status=%d %s", resp.StatusCode, string(b))
	}

	return nil
}
