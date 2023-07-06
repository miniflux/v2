// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/validator"

import (
	"miniflux.app/locale"
	"miniflux.app/model"
	"miniflux.app/storage"
)

// ValidateUserCreationWithPassword validates user creation with a password.
func ValidateUserCreationWithPassword(store *storage.Storage, request *model.UserCreationRequest) *ValidationError {
	if request.Username == "" {
		return NewValidationError("error.user_mandatory_fields")
	}

	if store.UserExists(request.Username) {
		return NewValidationError("error.user_already_exists")
	}

	if err := validatePassword(request.Password); err != nil {
		return err
	}

	return nil
}

// ValidateUserModification validates user modifications.
func ValidateUserModification(store *storage.Storage, userID int64, changes *model.UserModificationRequest) *ValidationError {
	if changes.Username != nil {
		if *changes.Username == "" {
			return NewValidationError("error.user_mandatory_fields")
		} else if store.AnotherUserExists(userID, *changes.Username) {
			return NewValidationError("error.user_already_exists")
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

func validateReadingSpeed(readingSpeed int) *ValidationError {
	if readingSpeed <= 0 {
		return NewValidationError("error.settings_reading_speed_is_positive")
	}
	return nil
}

func validatePassword(password string) *ValidationError {
	if len(password) < 6 {
		return NewValidationError("error.password_min_length")
	}
	return nil
}

func validateTheme(theme string) *ValidationError {
	themes := model.Themes()
	if _, found := themes[theme]; !found {
		return NewValidationError("error.invalid_theme")
	}
	return nil
}

func validateLanguage(language string) *ValidationError {
	languages := locale.AvailableLanguages()
	if _, found := languages[language]; !found {
		return NewValidationError("error.invalid_language")
	}
	return nil
}

func validateTimezone(store *storage.Storage, timezone string) *ValidationError {
	timezones, err := store.Timezones()
	if err != nil {
		return NewValidationError(err.Error())
	}

	if _, found := timezones[timezone]; !found {
		return NewValidationError("error.invalid_timezone")
	}
	return nil
}

func validateEntryDirection(direction string) *ValidationError {
	if direction != "asc" && direction != "desc" {
		return NewValidationError("error.invalid_entry_direction")
	}
	return nil
}

func validateEntriesPerPage(entriesPerPage int) *ValidationError {
	if entriesPerPage < 1 {
		return NewValidationError("error.entries_per_page_invalid")
	}
	return nil
}

func validateDisplayMode(displayMode string) *ValidationError {
	if displayMode != "fullscreen" && displayMode != "standalone" && displayMode != "minimal-ui" && displayMode != "browser" {
		return NewValidationError("error.invalid_display_mode")
	}
	return nil
}

func validateGestureNav(gestureNav string) *ValidationError {
	if gestureNav != "none" && gestureNav != "tap" && gestureNav != "swipe" {
		return NewValidationError("error.invalid_gesture_nav")
	}
	return nil
}

func validateDefaultHomePage(defaultHomePage string) *ValidationError {
	defaultHomePages := model.HomePages()
	if _, found := defaultHomePages[defaultHomePage]; !found {
		return NewValidationError("error.invalid_default_home_page")
	}
	return nil
}
