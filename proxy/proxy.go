// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"path"

	"miniflux.app/http/route"

	"github.com/gorilla/mux"

	"miniflux.app/config"
)

// ProxifyURL generates a relative URL for a proxified resource.
func ProxifyURL(router *mux.Router, link string) string {
	if link != "" {
		proxyImageUrl := config.Opts.ProxyImageUrl()

		if proxyImageUrl == "" {
			mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
			mac.Write([]byte(link))
			digest := mac.Sum(nil)
			return route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
		}

		proxyUrl, err := url.Parse(proxyImageUrl)
		if err != nil {
			return ""
		}

		proxyUrl.Path = path.Join(proxyUrl.Path, base64.URLEncoding.EncodeToString([]byte(link)))
		return proxyUrl.String()
	}
	return ""
}

// AbsoluteProxifyURL generates an absolute URL for a proxified resource.
func AbsoluteProxifyURL(router *mux.Router, host, link string) string {
	if link != "" {
		proxyImageUrl := config.Opts.ProxyImageUrl()

		if proxyImageUrl == "" {
			mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
			mac.Write([]byte(link))
			digest := mac.Sum(nil)
			path := route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
			if config.Opts.HTTPS {
				return "https://" + host + path
			} else {
				return "http://" + host + path
			}
		}

		proxyUrl, err := url.Parse(proxyImageUrl)
		if err != nil {
			return ""
		}

		proxyUrl.Path = path.Join(proxyUrl.Path, base64.URLEncoding.EncodeToString([]byte(link)))
		return proxyUrl.String()
	}
	return ""
}
