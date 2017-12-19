// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package wallabag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/miniflux/miniflux/http"
)

// Client represents a Wallabag client.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
}

// AddEntry sends a link to Wallabag.
func (c *Client) AddEntry(link, title string) error {
	accessToken, err := c.getAccessToken()
	if err != nil {
		return err
	}

	return c.createEntry(accessToken, link, title)
}

func (c *Client) createEntry(accessToken, link, title string) error {
	endpoint, err := getAPIEndpoint(c.baseURL, "/api/entries.json")
	if err != nil {
		return fmt.Errorf("wallbag: unable to get entries endpoint: %v", err)
	}

	client := http.NewClientWithAuthorization(endpoint, "Bearer "+accessToken)
	response, err := client.PostJSON(map[string]string{"url": link, "title": title})
	if err != nil {
		return fmt.Errorf("wallabag: unable to post entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("wallabag: request failed, status=%d", response.StatusCode)
	}

	return nil
}

func (c *Client) getAccessToken() (string, error) {
	values := url.Values{}
	values.Add("grant_type", "password")
	values.Add("client_id", c.clientID)
	values.Add("client_secret", c.clientSecret)
	values.Add("username", c.username)
	values.Add("password", c.password)

	endpoint, err := getAPIEndpoint(c.baseURL, "/oauth/v2/token")
	if err != nil {
		return "", fmt.Errorf("wallbag: unable to get token endpoint: %v", err)
	}

	client := http.NewClient(endpoint)
	response, err := client.PostForm(values)
	if err != nil {
		return "", fmt.Errorf("wallabag: unable to get access token: %v", err)
	}

	if response.HasServerFailure() {
		return "", fmt.Errorf("wallabag: request failed, status=%d", response.StatusCode)
	}

	token, err := decodeTokenResponse(response.Body)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// NewClient returns a new Wallabag client.
func NewClient(baseURL, clientID, clientSecret, username, password string) *Client {
	return &Client{baseURL, clientID, clientSecret, username, password}
}

func getAPIEndpoint(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("wallabag: invalid API endpoint: %v", err)
	}
	u.Path = path
	return u.String(), nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	Expires      int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func decodeTokenResponse(body io.Reader) (*tokenResponse, error) {
	var token tokenResponse

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&token); err != nil {
		return nil, fmt.Errorf("wallabag: unable to decode token response: %v", err)
	}

	return &token, nil
}
