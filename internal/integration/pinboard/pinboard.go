// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinboard // import "miniflux.app/v2/internal/integration/pinboard"

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	authToken string
}

func NewClient(authToken string) *Client {
	return &Client{authToken: authToken}
}

func (c *Client) CreateBookmark(entryURL, entryTitle, pinboardTags string, markAsUnread bool) error {
	if c.authToken == "" {
		return fmt.Errorf("pinboard: missing auth token")
	}

	toRead := "no"
	if markAsUnread {
		toRead = "yes"
	}

	values := url.Values{}
	values.Add("auth_token", c.authToken)
	values.Add("url", entryURL)
	values.Add("description", entryTitle)
	values.Add("tags", pinboardTags)
	values.Add("toread", toRead)

	apiEndpoint := "https://api.pinboard.in/v1/posts/add?" + values.Encode()
	request, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return fmt.Errorf("pinboard: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("pinboard: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("pinboard: unable to create a bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}
