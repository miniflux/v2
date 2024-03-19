// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxy // import "miniflux.app/v2/internal/proxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"net/url"
	"path"

	"miniflux.app/v2/internal/http/route"

	"github.com/gorilla/mux"

	"miniflux.app/v2/internal/config"
)

// ProxifyURL generates a relative URL for a proxified resource.
func ProxifyURL(router *mux.Router, mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.ProxyUrl(); customProxyURL != "" {
		return ProxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
	mac.Write([]byte(mediaURL))
	digest := mac.Sum(nil)
	return route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(mediaURL)))
}

// AbsoluteProxifyURL generates an absolute URL for a proxified resource.
func AbsoluteProxifyURL(router *mux.Router, host, mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.ProxyUrl(); customProxyURL != "" {
		return ProxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	proxifiedUrl := ProxifyURL(router, mediaURL)
	scheme := "http"
	if config.Opts.HTTPS {
		scheme = "https"
	}

	return scheme + "://" + host + proxifiedUrl
}

func ProxifyURLWithCustomProxy(mediaURL, customProxyURL string) string {
	if customProxyURL == "" {
		return mediaURL
	}

	proxyUrl, err := url.Parse(customProxyURL)
	if err != nil {
		slog.Error("Incorrect custom media proxy URL",
			slog.String("custom_proxy_url", customProxyURL),
			slog.Any("error", err),
		)
		return mediaURL
	}

	proxyUrl.Path = path.Join(proxyUrl.Path, base64.URLEncoding.EncodeToString([]byte(mediaURL)))
	return proxyUrl.String()
}
