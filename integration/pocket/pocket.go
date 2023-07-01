// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pocket // import "miniflux.app/integration/pocket"

import (
	"fmt"

	"miniflux.app/http/client"
)

// Client represents a Pocket client.
type Client struct {
	consumerKey string
	accessToken string
}

// NewClient returns a new Pocket client.
func NewClient(consumerKey, accessToken string) *Client {
	return &Client{consumerKey, accessToken}
}

// AddURL sends a single link to Pocket.
func (c *Client) AddURL(link, title string) error {
	if c.consumerKey == "" || c.accessToken == "" {
		return fmt.Errorf("pocket: missing credentials")
	}

	type body struct {
		AccessToken string `json:"access_token"`
		ConsumerKey string `json:"consumer_key"`
		Title       string `json:"title,omitempty"`
		URL         string `json:"url"`
	}

	data := &body{
		AccessToken: c.accessToken,
		ConsumerKey: c.consumerKey,
		Title:       title,
		URL:         link,
	}

	clt := client.New("https://getpocket.com/v3/add")
	response, err := clt.PostJSON(data)
	if err != nil {
		return fmt.Errorf("pocket: unable to send url: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("pocket: unable to send url, status=%d", response.StatusCode)
	}

	return nil
}
