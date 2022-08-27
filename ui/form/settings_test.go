package form // import "miniflux.app/ui/form"

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
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
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
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
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
		DefaultReadingSpeed: 35,
		CJKReadingSpeed:     25,
		DefaultHomePage:     "unread",
	}

	err := settings.Validate()
	if err == nil {
		t.Error("Validate should return an error")
	}
}
