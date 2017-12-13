// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter

import (
	"encoding/base64"
	"strings"

	"github.com/miniflux/miniflux/server/route"
	"github.com/miniflux/miniflux/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// ImageProxyFilter rewrites image tag URLs without HTTPS to local proxy URL
func ImageProxyFilter(router *mux.Router, data string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if srcAttr, ok := img.Attr("src"); ok {
			if !url.IsHTTPS(srcAttr) {
				img.SetAttr("src", Proxify(router, srcAttr))
			}
		}
	})

	output, _ := doc.Find("body").First().Html()
	return output
}

// Proxify returns a proxified link.
func Proxify(router *mux.Router, link string) string {
	return route.Path(router, "proxy", "encodedURL", base64.StdEncoding.EncodeToString([]byte(link)))
}
