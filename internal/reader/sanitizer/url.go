// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import "strings"

// validURISchemes is the allowlist for URLs in sanitized feed body content.
// It is intentionally broad; stricter surfaces (redirects, template hrefs)
// should use urllib.IsAbsoluteURL / urllib.IsRelativePath instead.
//
// See https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml
var validURISchemes = []string{
	// Most commong schemes on top.
	"https:",
	"http:",

	// Then the rest.
	"apt:",
	"bitcoin:",
	"callto:",
	"dav:",
	"davs:",
	"ed2k:",
	"facetime:",
	"feed:",
	"ftp:",
	"geo:",
	"git:",
	"gopher:",
	"irc:",
	"irc6:",
	"ircs:",
	"itms-apps:",
	"itms:",
	"magnet:",
	"mailto:",
	"news:",
	"nntp:",
	"rtmp:",
	"sftp:",
	"sip:",
	"sips:",
	"shortcuts:",
	"skype:",
	"spotify:",
	"ssh:",
	"steam:",
	"svn:",
	"svn+ssh:",
	"tel:",
	"webcal:",
	"xmpp:",

	// iOS Apps
	"opener:", // https://www.opener.link
	"hack:",   // https://apps.apple.com/it/app/hack-for-hacker-news-reader/id1464477788?l=en-GB
}

// HasValidURIScheme reports whether the URL begins with an allowed scheme.
func HasValidURIScheme(absoluteURL string) bool {
	for _, scheme := range validURISchemes {
		if strings.HasPrefix(absoluteURL, scheme) {
			return true
		}
	}
	return false
}
