// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Cubox API documentation: https://help.cubox.cc/save/api/

package cubox // import "miniflux.app/v2/internal/integration/cubox"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

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

	requestBody, err := json.Marshal(&card{
		Type:    "url",
		Content: entryURL,
	})
	if err != nil {
		return fmt.Errorf("cubox: unable to encode request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultClientTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiLink, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("cubox: unable to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("cubox: unable to send request: %w", err)
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
