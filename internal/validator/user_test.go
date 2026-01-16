// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/locale"
)

func TestValidateUsername(t *testing.T) {
	scenarios := map[string]*locale.LocalizedError{
		"userone":          nil,
		"user.name":        nil,
		"user@example.com": nil,
		"john_doe":         nil,
		"john-doe":         nil,
		"User123":          nil,
		"invalid username": locale.NewLocalizedError("error.invalid_username"),
		"user/path":        locale.NewLocalizedError("error.invalid_username"),
		"user🙂":            locale.NewLocalizedError("error.invalid_username"),
	}

	for username, expected := range scenarios {
		result := validateUsername(username)
		if expected == nil {
			if result != nil {
				t.Errorf(`got an unexpected error for %q instead of nil: %v`, username, result)
			}
		} else {
			if result == nil {
				t.Errorf(`expected an error, got nil.`)
			}
		}
	}
}

func TestValidateReadingSpeed(t *testing.T) {
	tests := map[int]bool{
		1:   false,
		100: false,
		0:   true,
		-5:  true,
	}

	for speed, wantErr := range tests {
		if err := validateReadingSpeed(speed); (err != nil) != wantErr {
			t.Errorf("reading speed %d error mismatch: got %v wantErr %v", speed, err, wantErr)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := map[string]bool{
		"secret":   false,
		"longpass": false,
		"short":    true,
		"":         true,
	}

	for password, wantErr := range tests {
		if err := validatePassword(password); (err != nil) != wantErr {
			t.Errorf("password %q error mismatch: got %v wantErr %v", password, err, wantErr)
		}
	}
}

func TestValidateTheme(t *testing.T) {
	if err := validateTheme("light_serif"); err != nil {
		t.Errorf("expected valid theme to pass, got %v", err)
	}

	if err := validateTheme("unknown"); err == nil {
		t.Error("expected invalid theme to fail")
	}
}

func TestValidateLanguage(t *testing.T) {
	if err := validateLanguage("en_US"); err != nil {
		t.Errorf("expected valid language to pass, got %v", err)
	}

	if err := validateLanguage("xx_YY"); err == nil {
		t.Error("expected invalid language to fail")
	}
}

func TestValidateTimezone(t *testing.T) {
	if err := validateTimezone("UTC"); err != nil {
		t.Errorf("expected valid timezone to pass, got %v", err)
	}

	if err := validateTimezone("Invalid/Zone"); err == nil {
		t.Error("expected invalid timezone to fail")
	}
}

func TestValidateEntryDirection(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		if err := validateEntryDirection(direction); err != nil {
			t.Errorf("expected valid direction %q to pass, got %v", direction, err)
		}
	}

	if err := validateEntryDirection("sideways"); err == nil {
		t.Error("expected invalid direction to fail")
	}
}

func TestValidateEntriesPerPage(t *testing.T) {
	if err := validateEntriesPerPage(1); err != nil {
		t.Errorf("expected positive entries per page to pass, got %v", err)
	}

	for _, value := range []int{0, -1} {
		if err := validateEntriesPerPage(value); err == nil {
			t.Errorf("expected %d to fail", value)
		}
	}
}

func TestValidateCategoriesSortingOrder(t *testing.T) {
	for _, order := range []string{"alphabetical", "unread_count"} {
		if err := validateCategoriesSortingOrder(order); err != nil {
			t.Errorf("expected valid order %q to pass, got %v", order, err)
		}
	}

	if err := validateCategoriesSortingOrder("popularity"); err == nil {
		t.Error("expected invalid order to fail")
	}
}

func TestValidateDisplayMode(t *testing.T) {
	for _, mode := range []string{"fullscreen", "standalone", "minimal-ui", "browser"} {
		if err := validateDisplayMode(mode); err != nil {
			t.Errorf("expected valid mode %q to pass, got %v", mode, err)
		}
	}

	if err := validateDisplayMode("windowed"); err == nil {
		t.Error("expected invalid display mode to fail")
	}
}

func TestValidateGestureNav(t *testing.T) {
	for _, gesture := range []string{"none", "tap", "swipe"} {
		if err := validateGestureNav(gesture); err != nil {
			t.Errorf("expected valid gesture %q to pass, got %v", gesture, err)
		}
	}

	if err := validateGestureNav("pinch"); err == nil {
		t.Error("expected invalid gesture to fail")
	}
}

func TestValidateDefaultHomePage(t *testing.T) {
	if err := validateDefaultHomePage("unread"); err != nil {
		t.Errorf("expected valid home page to pass, got %v", err)
	}

	if err := validateDefaultHomePage("dashboard"); err == nil {
		t.Error("expected invalid home page to fail")
	}
}

func TestValidateMediaPlaybackRate(t *testing.T) {
	for _, rate := range []float64{0.25, 1.0, 4.0} {
		if err := validateMediaPlaybackRate(rate); err != nil {
			t.Errorf("expected valid rate %.2f to pass, got %v", rate, err)
		}
	}

	for _, rate := range []float64{0.1, 4.1} {
		if err := validateMediaPlaybackRate(rate); err == nil {
			t.Errorf("expected invalid rate %.2f to fail", rate)
		}
	}
}
