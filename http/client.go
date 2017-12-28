// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/logger"
)

// Note: Some websites have a user agent filter.
const userAgent = "Mozilla/5.0 (like Gecko, like Safari, like Chrome) - Miniflux <https://miniflux.net/>"
const requestTimeout = 300

// Client is a HTTP Client :)
type Client struct {
	url                 string
	etagHeader          string
	lastModifiedHeader  string
	authorizationHeader string
	username            string
	password            string
	Insecure            bool
}

// Get execute a GET HTTP request.
func (c *Client) Get() (*Response, error) {
	request, err := c.buildRequest(http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest(request)
}

// PostForm execute a POST HTTP request with form values.
func (c *Client) PostForm(values url.Values) (*Response, error) {
	request, err := c.buildRequest(http.MethodPost, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.executeRequest(request)
}

// PostJSON execute a POST HTTP request with JSON payload.
func (c *Client) PostJSON(data interface{}) (*Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	request, err := c.buildRequest(http.MethodPost, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	return c.executeRequest(request)
}

func (c *Client) executeRequest(request *http.Request) (*Response, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[HttpClient] url=%s", c.url))

	client := c.buildClient()
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	response := &Response{
		Body:         resp.Body,
		StatusCode:   resp.StatusCode,
		EffectiveURL: resp.Request.URL.String(),
		LastModified: resp.Header.Get("Last-Modified"),
		ETag:         resp.Header.Get("ETag"),
		ContentType:  resp.Header.Get("Content-Type"),
	}

	logger.Debug("[HttpClient:%s] OriginalURL=%s, StatusCode=%d, ETag=%s, LastModified=%s, EffectiveURL=%s",
		request.Method,
		c.url,
		response.StatusCode,
		response.ETag,
		response.LastModified,
		response.EffectiveURL,
	)

	return response, err
}

func (c *Client) buildRequest(method string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, c.url, body)
	if err != nil {
		return nil, err
	}

	if c.username != "" && c.password != "" {
		request.SetBasicAuth(c.username, c.password)
	}

	request.Header = c.buildHeaders()
	return request, nil
}

func (c *Client) buildClient() http.Client {
	client := http.Client{Timeout: time.Duration(requestTimeout * time.Second)}
	if c.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

func (c *Client) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", userAgent)
	headers.Add("Accept", "*/*")

	if c.etagHeader != "" {
		headers.Add("If-None-Match", c.etagHeader)
	}

	if c.lastModifiedHeader != "" {
		headers.Add("If-Modified-Since", c.lastModifiedHeader)
	}

	if c.authorizationHeader != "" {
		headers.Add("Authorization", c.authorizationHeader)
	}

	return headers
}

// NewClient returns a new HTTP client.
func NewClient(url string) *Client {
	return &Client{url: url, Insecure: false}
}

// NewClientWithCredentials returns a new HTTP client that requires authentication.
func NewClientWithCredentials(url, username, password string) *Client {
	return &Client{url: url, Insecure: false, username: username, password: password}
}

// NewClientWithAuthorization returns a new client with a custom authorization header.
func NewClientWithAuthorization(url, authorization string) *Client {
	return &Client{url: url, Insecure: false, authorizationHeader: authorization}
}

// NewClientWithCacheHeaders returns a new HTTP client that send cache headers.
func NewClientWithCacheHeaders(url, etagHeader, lastModifiedHeader string) *Client {
	return &Client{url: url, etagHeader: etagHeader, lastModifiedHeader: lastModifiedHeader, Insecure: false}
}
