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

	"miniflux.app/v2/internal/crypto"
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

type errorResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type successResponse struct {
	Data struct {
		SaveUrl struct {
			Url             string `json:"url"`
			ClientRequestId string `json:"clientRequestId"`
		} `json:"saveUrl"`
	} `json:"data"`
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
				"clientRequestId": crypto.GenerateUUID(),
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

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("omnivore: failed to parse response: %s", err)
	}

	if resp.StatusCode >= 400 {
		var errResponse errorResponse
		if err = json.Unmarshal(b, &errResponse); err != nil {
			return fmt.Errorf("omnivore: failed to save URL: status=%d %s", resp.StatusCode, string(b))
		}
		return fmt.Errorf("omnivore: failed to save URL: status=%d %s", resp.StatusCode, errResponse.Errors[0].Message)
	}

	var successReponse successResponse
	if err = json.Unmarshal(b, &successReponse); err != nil {
		return fmt.Errorf("omnivore: failed to parse response, however the request appears successful, is the url correct?: status=%d %s", resp.StatusCode, string(b))
	}

	return nil
}
