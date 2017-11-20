// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter

import (
	"encoding/base64"
	"github.com/miniflux/miniflux2/reader/url"
	"github.com/miniflux/miniflux2/server/route"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// ImageProxyFilter rewrites image tag URLs without HTTPS to local proxy URL
func ImageProxyFilter(r *mux.Router, data string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if srcAttr, ok := img.Attr("src"); ok {
			if !url.IsHTTPS(srcAttr) {
				path := route.GetRoute(r, "proxy", "encodedURL", base64.StdEncoding.EncodeToString([]byte(srcAttr)))
				img.SetAttr("src", path)
			}
		}
	})

	output, _ := doc.Find("body").First().Html()
	return output
}
