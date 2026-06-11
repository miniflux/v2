// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Cubox API documentation: https://help.cubox.cc/save/api/

package cubox // import "miniflux.app/v2/internal/integration/cubox"

import (
	"errors"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/http/client"
)

type Client struct {
	apiLink string
}

func NewClient(apiLink string) *Client {
	return &Client{apiLink: apiLink}
}

func (c *Client) SaveLink(entryURL string) error {
	if c.apiLink == "" {
		return errors.New("cubox: missing API link")
	}

	response, err := client.NewRequestBuilder(c.apiLink).
		WithMethod(http.MethodPost).
		WithJSON(&card{
			Type:    "url",
			Content: entryURL,
		}).
		Do()
	if err != nil {
		return fmt.Errorf("cubox: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("cubox: unable to save link: status=%d", response.StatusCode)
	}

	return nil
}

type card struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}
