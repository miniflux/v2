// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/miniflux/miniflux2/helper"
)

const userAgent = "Miniflux <https://miniflux.net/>"
const requestTimeout = 300

// Client is a HTTP Client :)
type Client struct {
	url                string
	etagHeader         string
	lastModifiedHeader string
	Insecure           bool
}

// Get execute a GET HTTP request.
func (h *Client) Get() (*Response, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[HttpClient:Get] url=%s", h.url))
	u, _ := url.Parse(h.url)

	req := &http.Request{
		URL:    u,
		Method: http.MethodGet,
		Header: h.buildHeaders(),
	}

	client := h.buildClient()
	resp, err := client.Do(req)
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

	log.Println("[HttpClient:Get]",
		"OriginalURL:", h.url,
		"StatusCode:", response.StatusCode,
		"ETag:", response.ETag,
		"LastModified:", response.LastModified,
		"EffectiveURL:", response.EffectiveURL,
	)

	return response, err
}

func (h *Client) buildClient() http.Client {
	client := http.Client{Timeout: time.Duration(requestTimeout * time.Second)}
	if h.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

func (h *Client) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", userAgent)

	if h.etagHeader != "" {
		headers.Add("If-None-Match", h.etagHeader)
	}

	if h.lastModifiedHeader != "" {
		headers.Add("If-Modified-Since", h.lastModifiedHeader)
	}

	return headers
}

// NewClient returns a new HTTP client.
func NewClient(url string) *Client {
	return &Client{url: url, Insecure: false}
}

// NewClientWithCacheHeaders returns a new HTTP client that send cache headers.
func NewClientWithCacheHeaders(url, etagHeader, lastModifiedHeader string) *Client {
	return &Client{url: url, etagHeader: etagHeader, lastModifiedHeader: lastModifiedHeader, Insecure: false}
}
