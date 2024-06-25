// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinboard // import "miniflux.app/v2/internal/integration/pinboard"

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/version"
)

var errPostNotFound = fmt.Errorf("pinboard: post not found")
var errMissingCredentials = fmt.Errorf("pinboard: missing auth token")

const defaultClientTimeout = 10 * time.Second

type Client struct {
	authToken string
}

func NewClient(authToken string) *Client {
	return &Client{authToken: authToken}
}

func (c *Client) CreateBookmark(entryURL, entryTitle, pinboardTags string, markAsUnread bool) error {
	if c.authToken == "" {
		return errMissingCredentials
	}

	// We check if the url is already bookmarked to avoid overriding existing data.
	post, err := c.getBookmark(entryURL)

	if err != nil && errors.Is(err, errPostNotFound) {
		post = NewPost(entryURL, entryTitle)
	} else if err != nil {
		// In case of any other error, we return immediately to avoid overriding existing data.
		return err
	}

	post.addTag(pinboardTags)
	if markAsUnread {
		post.SetToread()
	}

	values := url.Values{}
	values.Add("auth_token", c.authToken)
	post.AddValues(values)

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

// getBookmark fetches a bookmark from Pinboard. https://www.pinboard.in/api/#posts_get
func (c *Client) getBookmark(entryURL string) (*Post, error) {
	if c.authToken == "" {
		return nil, errMissingCredentials
	}

	values := url.Values{}
	values.Add("auth_token", c.authToken)
	values.Add("url", entryURL)

	apiEndpoint := "https://api.pinboard.in/v1/posts/get?" + values.Encode()
	request, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("pinboard: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("pinboard: unable fetch bookmark: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("pinboard: unable to fetch bookmark, status=%d", response.StatusCode)
	}

	var results posts
	err = xml.NewDecoder(response.Body).Decode(&results)
	if err != nil {
		return nil, fmt.Errorf("pinboard: unable to decode XML: %v", err)
	}

	if len(results.Posts) == 0 {
		return nil, errPostNotFound
	}

	return &results.Posts[0], nil
}
