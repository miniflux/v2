// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"

	"github.com/gorilla/mux"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/route"
)

func ProxifyRelativeURL(router *mux.Router, mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != nil {
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

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != nil {
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

func proxifyURLWithCustomProxy(mediaURL string, customProxyURL *url.URL) string {
	if customProxyURL == nil {
		return mediaURL
	}

	absoluteURL := customProxyURL.JoinPath(base64.URLEncoding.EncodeToString([]byte(mediaURL)))
	return absoluteURL.String()
}
