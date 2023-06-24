// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package wallabag // import "miniflux.app/integration/wallabag"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"miniflux.app/http/client"
)

// Client represents a Wallabag client.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
	onlyURL      bool
}

// NewClient returns a new Wallabag client.
func NewClient(baseURL, clientID, clientSecret, username, password string, onlyURL bool) *Client {
	return &Client{baseURL, clientID, clientSecret, username, password, onlyURL}
}

// AddEntry sends a link to Wallabag.
// Pass an empty string in `content` to let Wallabag fetch the article content.
func (c *Client) AddEntry(link, title, content string) error {
	if c.baseURL == "" || c.clientID == "" || c.clientSecret == "" || c.username == "" || c.password == "" {
		return fmt.Errorf("wallabag: missing credentials")
	}

	accessToken, err := c.getAccessToken()
	if err != nil {
		return err
	}

	return c.createEntry(accessToken, link, title, content)
}

func (c *Client) createEntry(accessToken, link, title, content string) error {
	endpoint, err := url.JoinPath(c.baseURL, "/api/entries.json")
	if err != nil {
		return fmt.Errorf("wallbag: unable to get entries endpoint: %v", err)
	}

	data := map[string]string{"url": link, "title": title}
	if !c.onlyURL {
		data["content"] = content
	}

	clt := client.New(endpoint)
	clt.WithAuthorization("Bearer " + accessToken)
	response, err := clt.PostJSON(data)
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

	endpoint, err := url.JoinPath(c.baseURL, "/oauth/v2/token")
	if err != nil {
		return "", fmt.Errorf("wallbag: unable to get token endpoint: %v", err)
	}

	clt := client.New(endpoint)
	response, err := clt.PostForm(values)
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
