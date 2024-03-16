// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package shaarli // import "miniflux.app/v2/internal/integration/shaarli"

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL   string
	apiSecret string
}

func NewClient(baseURL, apiSecret string) *Client {
	return &Client{baseURL: baseURL, apiSecret: apiSecret}
}

func (c *Client) CreateLink(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiSecret == "" {
		return fmt.Errorf("shaarli: missing base URL or API secret")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/v1/links")
	if err != nil {
		return fmt.Errorf("shaarli: invalid API endpoint: %v", err)
	}

	requestBody, err := json.Marshal(&addLinkRequest{
		URL:     entryURL,
		Title:   entryTitle,
		Private: true,
	})

	if err != nil {
		return fmt.Errorf("shaarli: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("shaarli: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.generateBearerToken())

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("shaarli: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("shaarli: unable to add link: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

func (c *Client) generateBearerToken() string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"typ":"JWT","alg":"HS512"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"iat":%d}`, time.Now().Unix())))
	data := header + "." + payload

	mac := hmac.New(sha512.New, []byte(c.apiSecret))
	mac.Write([]byte(data))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return data + "." + signature
}

type addLinkRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Private bool   `json:"private"`
}
