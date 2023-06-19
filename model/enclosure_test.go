// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"
)

func TestEnclosure_Html5MimeTypeGivesOriginalMimeType(t *testing.T) {
	enclosure := Enclosure{MimeType: "thing/thisMimeTypeIsNotExpectedToBeReplaced"}
	if enclosure.Html5MimeType() != enclosure.MimeType {
		t.Fatalf(
			"HTML5 MimeType must provide original MimeType if not explicitly Replaced. Got %s ,expected '%s' ",
			enclosure.Html5MimeType(),
			enclosure.MimeType,
		)
	}
}

func TestEnclosure_Html5MimeTypeReplaceStandardM4vByAppleSpecificMimeType(t *testing.T) {
	enclosure := Enclosure{MimeType: "video/m4v"}
	if enclosure.Html5MimeType() != "video/x-m4v" {
		// Solution from this stackoverflow discussion:
		// https://stackoverflow.com/questions/15277147/m4v-mimetype-video-mp4-or-video-m4v/66945470#66945470
		// tested at the time of this commit (06/2023) on latest Firefox & Vivaldi on this feed
		// https://www.florenceporcel.com/podcast/lfhdu.xml
		t.Fatalf(
			"HTML5 MimeType must be replaced by 'video/x-m4v' when originally video/m4v to ensure playbacks in brownser. Got '%s'",
			enclosure.Html5MimeType(),
		)
	}
}
