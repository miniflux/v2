// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter

import (
	"encoding/base64"
	"strings"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// ImageProxyFilter rewrites image tag URLs to local proxy URL (by default only non-HTTPS URLs)
func ImageProxyFilter(router *mux.Router, cfg *config.Config, data string) string {
	proxyImages := cfg.ProxyImages()
	if proxyImages == 0 {
		return data
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if srcAttr, ok := img.Attr("src"); ok {
			if proxyImages == 2 || !url.IsHTTPS(srcAttr) {
				img.SetAttr("src", Proxify(router, srcAttr))
			}
		}
	})

	output, _ := doc.Find("body").First().Html()
	return output
}

// Proxify returns a proxified link.
func Proxify(router *mux.Router, link string) string {
	// We use base64 url encoding to avoid slash in the URL.
	return route.Path(router, "proxy", "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
}
