// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"

	"miniflux.app/v2/internal/config"
)

func ProxifyRelativeURL(mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != nil {
		return proxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	mediaURLBytes := []byte(mediaURL)

	mac := hmac.New(sha256.New, config.Opts.MediaProxyPrivateKey())
	mac.Write(mediaURLBytes)
	digest := mac.Sum(nil)

	// Preserve the configured base path so proxied URLs still work when Miniflux is served from a subfolder.
	return fmt.Sprintf("%s/proxy/%s/%s", config.Opts.BasePath(), base64.URLEncoding.EncodeToString(digest), base64.URLEncoding.EncodeToString(mediaURLBytes))
}

func ProxifyAbsoluteURL(mediaURL string) string {
	if mediaURL == "" {
		return ""
	}

	if customProxyURL := config.Opts.MediaCustomProxyURL(); customProxyURL != nil {
		return proxifyURLWithCustomProxy(mediaURL, customProxyURL)
	}

	// Note that the proxyified URL is relative to the root URL.
	proxifiedUrl := ProxifyRelativeURL(mediaURL)
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
