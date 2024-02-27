// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
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
		html.BadRequest(w, r, errors.New("no URL provided"))
		return
	}

	decodedDigest, err := base64.URLEncoding.DecodeString(encodedDigest)
	if err != nil {
		html.BadRequest(w, r, errors.New("unable to decode this digest"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("unable to decode this URL"))
		return
	}

	mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
	mac.Write(decodedURL)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decodedDigest, expectedMAC) {
		html.Forbidden(w, r)
		return
	}

	u, err := url.Parse(string(decodedURL))
	if err != nil {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if u.Host == "" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if !u.IsAbs() {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	mediaURL := string(decodedURL)
	slog.Debug("MediaProxy: Fetching remote resource",
		slog.String("media_url", mediaURL),
	)

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
		slog.Error("MediaProxy: Unable to initialize HTTP client",
			slog.String("media_url", mediaURL),
			slog.Any("error", err),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		slog.Warn("MediaProxy: "+http.StatusText(http.StatusRequestedRangeNotSatisfiable),
			slog.String("media_url", mediaURL),
			slog.Int("status_code", resp.StatusCode),
		)
		html.RequestedRangeNotSatisfiable(w, r, resp.Header.Get("Content-Range"))
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		slog.Warn("MediaProxy: Unexpected response status code",
			slog.String("media_url", mediaURL),
			slog.Int("status_code", resp.StatusCode),
		)
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
