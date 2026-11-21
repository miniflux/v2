// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"testing"
)

func TestValid(t *testing.T) {
	settings := &SettingsForm{
		Theme:                   "default",
		Language:                "en_US",
		Timezone:                "UTC",
		EntryDirection:          "asc",
		EntriesPerPage:          50,
		DisplayMode:             "standalone",
		GestureNav:              "tap",
		DefaultReadingSpeed:     35,
		CJKReadingSpeed:         25,
		DefaultHomePage:         "unread",
		MediaPlaybackRate:       1.25,
		AlwaysOpenExternalLinks: true,
	}

	err := settings.Validate()
	if err != nil {
		t.Error(err)
	}
}
