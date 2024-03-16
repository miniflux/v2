// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"testing"
)

func TestValid(t *testing.T) {
	settings := &SettingsForm{
		Username:            "user",
		Password:            "hunter2",
		Confirmation:        "hunter2",
		Theme:               "default",
		Language:            "en_US",
		Timezone:            "UTC",
		EntryDirection:      "asc",
		EntriesPerPage:      50,
		DisplayMode:         "standalone",
		GestureNav:          "tap",
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
		MediaPlaybackRate:   1.25,
	}

	err := settings.Validate()
	if err != nil {
		t.Error(err)
	}
}

func TestConfirmationEmpty(t *testing.T) {
	settings := &SettingsForm{
		Username:            "user",
		Password:            "hunter2",
		Confirmation:        "",
		Theme:               "default",
		Language:            "en_US",
		Timezone:            "UTC",
		EntryDirection:      "asc",
		EntriesPerPage:      50,
		DisplayMode:         "standalone",
		GestureNav:          "tap",
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
		MediaPlaybackRate:   1.25,
	}

	err := settings.Validate()
	if err != nil {
		t.Error(err)
	}

	if settings.Password != "" {
		t.Error("Password should have been cleared")
	}
}

func TestConfirmationIncorrect(t *testing.T) {
	settings := &SettingsForm{
		Username:            "user",
		Password:            "hunter2",
		Confirmation:        "unter2",
		Theme:               "default",
		Language:            "en_US",
		Timezone:            "UTC",
		EntryDirection:      "asc",
		EntriesPerPage:      50,
		DisplayMode:         "standalone",
		GestureNav:          "tap",
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
		MediaPlaybackRate:   1.25,
	}

	err := settings.Validate()
	if err == nil {
		t.Error("Validate should return an error")
	}
}
