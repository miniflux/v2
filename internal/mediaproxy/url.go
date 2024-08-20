// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"net/url"

	"github.com/gorilla/mux"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/route"
)

func ProxifyRelativeURL(router *mux.Router, mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != "" {
		return proxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	mac := hmac.New(sha256.New, config.Opts.MediaProxyPrivateKey())
	mac.Write([]byte(mediaURL))
	digest := mac.Sum(nil)
	return route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(mediaURL)))
}

func ProxifyAbsoluteURL(router *mux.Router, mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != "" {
		return proxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	// Note that the proxyified URL is relative to the root URL.
	proxifiedUrl := ProxifyRelativeURL(router, mediaURL)
	absoluteURL, err := url.JoinPath(config.Opts.RootURL(), proxifiedUrl)
	if err != nil {
		return mediaURL
	}

	return absoluteURL
}

func proxifyURLWithCustomProxy(mediaURL, customProxyURL string) string {
	if customProxyURL == "" {
		return mediaURL
	}

	absoluteURL, err := url.JoinPath(customProxyURL, base64.URLEncoding.EncodeToString([]byte(mediaURL)))
	if err != nil {
		slog.Error("Incorrect custom media proxy URL",
			slog.String("custom_proxy_url", customProxyURL),
			slog.Any("error", err),
		)
		return mediaURL
	}

	return absoluteURL
}
