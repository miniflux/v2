// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package archiveorg

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 30 * time.Second

// See https://docs.google.com/document/d/1Nsv52MvSjbLb2PCpHlat0gkzw0EvtSgpKHu4mk0MnrA/edit?tab=t.0
const options = "delay_wb_availability=1&if_not_archived_within=15d"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SendURL(entryURL string) error {
	requestURL := "https://web.archive.org/save/" + url.QueryEscape(entryURL) + "?" + options
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("archiveorg: unable to create request: %v", err)
	}

	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := client.NewClientWithOptions(client.Options{Timeout: defaultClientTimeout})
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("archiveorg: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("archiveorg: unexpected status code: url=%s status=%d", requestURL, response.StatusCode)
	}

	return nil
}
