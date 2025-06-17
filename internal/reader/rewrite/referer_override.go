// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"net/url"
	"strings"
)

// GetRefererForURL returns the referer for the given URL if it exists, otherwise an empty string.
func GetRefererForURL(u string) string {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return ""
	}

	switch parsedUrl.Hostname() {
	case "appinn.com":
		return "https://appinn.com"
	case "bjp.org.cn":
		return "https://bjp.org.cn"
	case "cdnfile.sspai.com":
		return "https://sspai.com"
	case "f.video.weibocdn.com":
		return "https://weibo.com"
	case "i.pximg.net":
		return "https://www.pixiv.net"
	case "img.hellogithub.com":
		return "https://hellogithub.com"
	case "moyu.im":
		return "https://i.jandan.net"
	case "www.parkablogs.com":
		return "https://www.parkablogs.com"
	}

	switch {
	case strings.HasSuffix(parsedUrl.Hostname(), ".cdninstagram.com"):
		return "https://www.instagram.com"
	case strings.HasSuffix(parsedUrl.Hostname(), ".moyu.im"):
		return "https://i.jandan.net"
	case strings.HasSuffix(parsedUrl.Hostname(), ".sinaimg.cn"):
		return "https://weibo.com"
	}

	return ""
}
