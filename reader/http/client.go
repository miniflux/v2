// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"log"
	"net/http"
	"net/url"
	"time"
)

const HTTP_USER_AGENT = "Miniflux <https://miniflux.net/>"

type HttpClient struct {
	url                string
	etagHeader         string
	lastModifiedHeader string
	Insecure           bool
}

func (h *HttpClient) Get() (*ServerResponse, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[HttpClient:Get] url=%s", h.url))
	u, _ := url.Parse(h.url)

	req := &http.Request{
		URL:    u,
		Method: "GET",
		Header: h.buildHeaders(),
	}

	client := h.buildClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	response := &ServerResponse{
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

func (h *HttpClient) buildClient() http.Client {
	if h.Insecure {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		return http.Client{Transport: transport}
	}

	return http.Client{}
}

func (h *HttpClient) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", HTTP_USER_AGENT)

	if h.etagHeader != "" {
		headers.Add("If-None-Match", h.etagHeader)
	}

	if h.lastModifiedHeader != "" {
		headers.Add("If-Modified-Since", h.lastModifiedHeader)
	}

	return headers
}

func NewHttpClient(url string) *HttpClient {
	return &HttpClient{url: url, Insecure: false}
}

func NewHttpClientWithCacheHeaders(url, etagHeader, lastModifiedHeader string) *HttpClient {
	return &HttpClient{url: url, etagHeader: etagHeader, lastModifiedHeader: lastModifiedHeader, Insecure: false}
}
