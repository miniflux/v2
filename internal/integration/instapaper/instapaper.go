// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package instapaper // import "miniflux.app/v2/internal/integration/instapaper"

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	username string
	password string
}

func NewClient(username, password string) *Client {
	return &Client{username: username, password: password}
}

func (c *Client) AddURL(entryURL, entryTitle string) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("instapaper: missing username or password")
	}

	values := url.Values{}
	values.Add("url", entryURL)
	values.Add("title", entryTitle)

	apiEndpoint := "https://www.instapaper.com/api/add?" + values.Encode()
	request, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return fmt.Errorf("instapaper: unable to create request: %v", err)
	}

	request.SetBasicAuth(c.username, c.password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("instapaper: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("instapaper: unable to add URL: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}
