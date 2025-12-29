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
	"path"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/urllib"
)

func (h *handler) mediaProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the media is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("no URL provided"))
		return
	}

	encodedDigest := request.RouteStringParam(r, "encodedDigest")
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

	mac := hmac.New(sha256.New, config.Opts.MediaProxyPrivateKey())
	mac.Write(decodedURL)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decodedDigest, expectedMAC) {
		html.Forbidden(w, r)
		return
	}

	parsedMediaURL, err := url.Parse(string(decodedURL))
	if err != nil {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if parsedMediaURL.Scheme != "http" && parsedMediaURL.Scheme != "https" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if parsedMediaURL.Host == "" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if !parsedMediaURL.IsAbs() {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	mediaURL := string(decodedURL)

	if !config.Opts.MediaProxyAllowPrivateNetworks() {
		if isPrivate, err := urllib.ResolvesToPrivateIP(parsedMediaURL.Hostname()); err != nil {
			slog.Warn("MediaProxy: Unable to resolve hostname",
				slog.String("media_url", mediaURL),
				slog.Any("error", err),
			)
			html.Forbidden(w, r)
			return
		} else if isPrivate {
			slog.Warn("MediaProxy: Refusing to access private IP address",
				slog.String("media_url", mediaURL),
			)
			html.Forbidden(w, r)
			return
		}
	}

	slog.Debug("MediaProxy: Fetching remote resource",
		slog.String("media_url", mediaURL),
	)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.MediaProxyHTTPClientTimeout())

	// Disable compression for the media proxy requests (not implemented).
	requestBuilder.WithoutCompression()

	if referer := rewrite.GetRefererForURL(mediaURL); referer != "" {
		requestBuilder.WithHeader("Referer", referer)
	}

	forwardedRequestHeader := [...]string{"Range", "Accept", "Accept-Encoding", "User-Agent"}
	for _, requestHeaderName := range forwardedRequestHeader {
		if r.Header.Get(requestHeaderName) != "" {
			requestBuilder.WithHeader(requestHeaderName, r.Header.Get(requestHeaderName))
		}
	}

	resp, err := requestBuilder.ExecuteRequest(mediaURL)
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

		// Forward the status code from the origin.
		http.Error(w, "Origin status code is "+strconv.Itoa(resp.StatusCode), resp.StatusCode)
		return
	}

	etag := crypto.HashFromBytes(decodedURL)

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithStatus(resp.StatusCode)
		b.WithHeader("Content-Security-Policy", response.ContentSecurityPolicyForUntrustedContent)
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))

		if filename := path.Base(parsedMediaURL.Path); filename != "" {
			b.WithHeader("Content-Disposition", `inline; filename="`+filename+`"`)
		}

		forwardedResponseHeader := [...]string{"Content-Encoding", "Content-Type", "Content-Length", "Accept-Ranges", "Content-Range"}
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
