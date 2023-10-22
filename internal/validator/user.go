// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

// ValidateUserCreationWithPassword validates user creation with a password.
func ValidateUserCreationWithPassword(store *storage.Storage, request *model.UserCreationRequest) *locale.LocalizedError {
	if request.Username == "" {
		return locale.NewLocalizedError("error.user_mandatory_fields")
	}

	if store.UserExists(request.Username) {
		return locale.NewLocalizedError("error.user_already_exists")
	}

	if err := validatePassword(request.Password); err != nil {
		return err
	}

	return nil
}

// ValidateUserModification validates user modifications.
func ValidateUserModification(store *storage.Storage, userID int64, changes *model.UserModificationRequest) *locale.LocalizedError {
	if changes.Username != nil {
		if *changes.Username == "" {
			return locale.NewLocalizedError("error.user_mandatory_fields")
		} else if store.AnotherUserExists(userID, *changes.Username) {
			return locale.NewLocalizedError("error.user_already_exists")
		}
	}

	if changes.Password != nil {
		if err := validatePassword(*changes.Password); err != nil {
			return err
		}
	}

	if changes.Theme != nil {
		if err := validateTheme(*changes.Theme); err != nil {
			return err
		}
	}

	if changes.Language != nil {
		if err := validateLanguage(*changes.Language); err != nil {
			return err
		}
	}

	if changes.Timezone != nil {
		if err := validateTimezone(store, *changes.Timezone); err != nil {
			return err
		}
	}

	if changes.EntryDirection != nil {
		if err := validateEntryDirection(*changes.EntryDirection); err != nil {
			return err
		}
	}

	if changes.EntriesPerPage != nil {
		if err := validateEntriesPerPage(*changes.EntriesPerPage); err != nil {
			return err
		}
	}

	if changes.DisplayMode != nil {
		if err := validateDisplayMode(*changes.DisplayMode); err != nil {
			return err
		}
	}

	if changes.GestureNav != nil {
		if err := validateGestureNav(*changes.GestureNav); err != nil {
			return err
		}
	}

	if changes.DefaultReadingSpeed != nil {
		if err := validateReadingSpeed(*changes.DefaultReadingSpeed); err != nil {
			return err
		}
	}

	if changes.CJKReadingSpeed != nil {
		if err := validateReadingSpeed(*changes.CJKReadingSpeed); err != nil {
			return err
		}
	}

	if changes.DefaultHomePage != nil {
		if err := validateDefaultHomePage(*changes.DefaultHomePage); err != nil {
			return err
		}
	}

	return nil
}

func validateReadingSpeed(readingSpeed int) *locale.LocalizedError {
	if readingSpeed <= 0 {
		return locale.NewLocalizedError("error.settings_reading_speed_is_positive")
	}
	return nil
}

func validatePassword(password string) *locale.LocalizedError {
	if len(password) < 6 {
		return locale.NewLocalizedError("error.password_min_length")
	}
	return nil
}

func validateTheme(theme string) *locale.LocalizedError {
	themes := model.Themes()
	if _, found := themes[theme]; !found {
		return locale.NewLocalizedError("error.invalid_theme")
	}
	return nil
}

func validateLanguage(language string) *locale.LocalizedError {
	languages := locale.AvailableLanguages()
	if _, found := languages[language]; !found {
		return locale.NewLocalizedError("error.invalid_language")
	}
	return nil
}

func validateTimezone(store *storage.Storage, timezone string) *locale.LocalizedError {
	timezones, err := store.Timezones()
	if err != nil {
		return locale.NewLocalizedError(err.Error())
	}

	if _, found := timezones[timezone]; !found {
		return locale.NewLocalizedError("error.invalid_timezone")
	}
	return nil
}

func validateEntryDirection(direction string) *locale.LocalizedError {
	if direction != "asc" && direction != "desc" {
		return locale.NewLocalizedError("error.invalid_entry_direction")
	}
	return nil
}

func validateEntriesPerPage(entriesPerPage int) *locale.LocalizedError {
	if entriesPerPage < 1 {
		return locale.NewLocalizedError("error.entries_per_page_invalid")
	}
	return nil
}

func validateDisplayMode(displayMode string) *locale.LocalizedError {
	if displayMode != "fullscreen" && displayMode != "standalone" && displayMode != "minimal-ui" && displayMode != "browser" {
		return locale.NewLocalizedError("error.invalid_display_mode")
	}
	return nil
}

func validateGestureNav(gestureNav string) *locale.LocalizedError {
	if gestureNav != "none" && gestureNav != "tap" && gestureNav != "swipe" {
		return locale.NewLocalizedError("error.invalid_gesture_nav")
	}
	return nil
}

func validateDefaultHomePage(defaultHomePage string) *locale.LocalizedError {
	defaultHomePages := model.HomePages()
	if _, found := defaultHomePages[defaultHomePage]; !found {
		return locale.NewLocalizedError("error.invalid_default_home_page")
	}
	return nil
}
