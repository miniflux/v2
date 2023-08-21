// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/logger"
)

func (h *handler) mediaProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the media is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedDigest := request.RouteStringParam(r, "encodedDigest")
	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("No URL provided"))
		return
	}

	decodedDigest, err := base64.URLEncoding.DecodeString(encodedDigest)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this Digest"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this URL"))
		return
	}

	mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
	mac.Write(decodedURL)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decodedDigest, expectedMAC) {
		html.Forbidden(w, r)
		return
	}

	mediaURL := string(decodedURL)
	logger.Debug(`[Proxy] Fetching %q`, mediaURL)

	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Note: User-Agent HTTP header is omitted to avoid being blocked by bot protection mechanisms.
	req.Header.Add("Connection", "close")

	forwardedRequestHeader := []string{"Range", "Accept", "Accept-Encoding"}
	for _, requestHeaderName := range forwardedRequestHeader {
		if r.Header.Get(requestHeaderName) != "" {
			req.Header.Add(requestHeaderName, r.Header.Get(requestHeaderName))
		}
	}

	clt := &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: time.Duration(config.Opts.ProxyHTTPClientTimeout()) * time.Second,
		},
		Timeout: time.Duration(config.Opts.ProxyHTTPClientTimeout()) * time.Second,
	}

	resp, err := clt.Do(req)
	if err != nil {
		logger.Error(`[Proxy] Unable to initialize HTTP client: %v`, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		logger.Error(`[Proxy] Status Code is %d for URL %q`, resp.StatusCode, mediaURL)
		html.RequestedRangeNotSatisfiable(w, r, resp.Header.Get("Content-Range"))
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		logger.Error(`[Proxy] Status Code is %d for URL %q`, resp.StatusCode, mediaURL)
		html.NotFound(w, r)
		return
	}

	etag := crypto.HashFromBytes(decodedURL)

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithStatus(resp.StatusCode)
		b.WithHeader("Content-Security-Policy", `default-src 'self'`)
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))
		forwardedResponseHeader := []string{"Content-Encoding", "Content-Type", "Content-Length", "Accept-Ranges", "Content-Range"}
		for _, responseHeaderName := range forwardedResponseHeader {
			if resp.Header.Get(responseHeaderName) != "" {
				b.WithHeader(responseHeaderName, resp.Header.Get(responseHeaderName))
			}
		}
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}
