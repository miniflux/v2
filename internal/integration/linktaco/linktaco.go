// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linktaco // import "miniflux.app/v2/internal/integration/linktaco"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/version"
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultGraphQLURL    = "https://api.linktaco.com/query"
	maxTags              = 10
	maxDescriptionLength = 500
)

type Client struct {
	graphqlURL string
	apiToken   string
	orgSlug    string
	tags       string
	visibility string
}

func NewClient(apiToken, orgSlug, tags, visibility string) *Client {
	if visibility == "" {
		visibility = "PUBLIC"
	}
	return &Client{
		graphqlURL: defaultGraphQLURL,
		apiToken:   apiToken,
		orgSlug:    orgSlug,
		tags:       tags,
		visibility: visibility,
	}
}

func (c *Client) CreateBookmark(entryURL, entryTitle, entryContent string) error {
	if c.apiToken == "" || c.orgSlug == "" {
		return errors.New("linktaco: missing API token or organization slug")
	}

	description := entryContent
	if len(description) > maxDescriptionLength {
		description = description[:maxDescriptionLength]
	}

	// tags (limit to 10)
	tags := strings.FieldsFunc(c.tags, func(c rune) bool {
		return c == ',' || c == ' '
	})
	if len(tags) > maxTags {
		tags = tags[:maxTags]
	}
	// tagsStr is used in GraphQL query to pass comma separated tags
	tagsStr := strings.Join(tags, ",")

	mutation := `
		mutation AddLink($input: LinkInput!) {
			addLink(input: $input) {
				id
				url
				title
			}
		}
	`

	variables := map[string]any{
		"input": map[string]any{
			"url":         entryURL,
			"title":       entryTitle,
			"description": description,
			"orgSlug":     c.orgSlug,
			"visibility":  c.visibility,
			"unread":      true,
			"starred":     false,
			"archive":     false,
			"tags":        tagsStr,
		},
	}

	requestBody, err := json.Marshal(map[string]any{
		"query":     mutation,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("linktaco: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, c.graphqlURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("linktaco: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("linktaco: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("linktaco: unable to create bookmark: status=%d", response.StatusCode)
	}

	var graphqlResponse struct {
		Data   json.RawMessage   `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}

	if err := json.NewDecoder(response.Body).Decode(&graphqlResponse); err != nil {
		return fmt.Errorf("linktaco: unable to decode response: %v", err)
	}

	if len(graphqlResponse.Errors) > 0 {
		// Try to extract error message
		var errorMsg string
		for _, errJSON := range graphqlResponse.Errors {
			var errObj struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(errJSON, &errObj) == nil && errObj.Message != "" {
				errorMsg = errObj.Message
				break
			}
		}
		if errorMsg == "" {
			// Fallback. Should never be reached.
			errorMsg = "GraphQL error occurred (fallback message)"
		}
		return fmt.Errorf("linktaco: %s", errorMsg)
	}

	return nil
}
