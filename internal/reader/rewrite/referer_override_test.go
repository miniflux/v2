// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"testing"
)

func TestGetRefererForURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Weibo Image URL",
			url:      "https://wx1.sinaimg.cn/large/example.jpg",
			expected: "https://weibo.com",
		},
		{
			name:     "Pixiv Image URL",
			url:      "https://i.pximg.net/img-master/example.jpg",
			expected: "https://www.pixiv.net",
		},
		{
			name:     "SSPai CDN URL",
			url:      "https://cdnfile.sspai.com/example.png",
			expected: "https://sspai.com",
		},
		{
			name:     "Instagram CDN URL",
			url:      "https://scontent-sjc3-1.cdninstagram.com/example.jpg",
			expected: "https://www.instagram.com",
		},
		{
			name:     "Weibo Video URL",
			url:      "https://f.video.weibocdn.com/example.mp4",
			expected: "https://weibo.com",
		},
		{
			name:     "HelloGithub Image URL",
			url:      "https://img.hellogithub.com/example.png",
			expected: "https://hellogithub.com",
		},
		{
			name:     "Park Blogs",
			url:      "https://www.parkablogs.com/sites/default/files/2025/image.jpg",
			expected: "https://www.parkablogs.com",
		},
		{
			name:     "Non-matching URL",
			url:      "https://example.com/image.jpg",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetRefererForURL(tc.url)
			if result != tc.expected {
				t.Errorf("GetRefererForURL(%s): expected %s, got %s",
					tc.url, tc.expected, result)
			}
		})
	}
}
