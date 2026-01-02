// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var refererMappings = map[string]string{
	"appinn.com":           "https://appinn.com",
	"bjp.org.cn":           "https://bjp.org.cn",
	"cdnfile.sspai.com":    "https://sspai.com",
	"f.video.weibocdn.com": "https://weibo.com",
	"i.pximg.net":          "https://www.pixiv.net",
	"img.hellogithub.com":  "https://hellogithub.com",
	"moyu.im":              "https://i.jandan.net",
	"www.parkablogs.com":   "https://www.parkablogs.com",
	".cdninstagram.com":    "https://www.instagram.com",
	".moyu.im":             "https://i.jandan.net",
	".sinaimg.cn":          "https://weibo.com",
}

func LoadRefererOverrides(overridesString string) error {
	pairs := strings.Split(overridesString, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid referer override format: %q (expected domain=referer)", pair)
		}
		domain := strings.TrimSpace(parts[0])
		referer := strings.TrimSpace(parts[1])
		if domain == "" || referer == "" {
			return errors.New("invalid referer override: domain and referer cannot be empty")
		}
		refererMappings[domain] = referer
	}
	return nil
}

// GetRefererForURL returns the referer for the given URL if it exists, otherwise an empty string.
func GetRefererForURL(u string) string {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return ""
	}

	hostname := parsedUrl.Hostname()

	if referer, ok := refererMappings[hostname]; ok {
		return referer
	}

	for suffix, referer := range refererMappings {
		if strings.HasPrefix(suffix, ".") && strings.HasSuffix(hostname, suffix) {
			return referer
		}
	}

	return ""
}
