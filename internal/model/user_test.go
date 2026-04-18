// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"
)

func TestUserModificationRequestPatch(t *testing.T) {
	user := &User{
		Username:                         "original",
		Theme:                            "original_theme",
		Language:                         "original_language",
		Timezone:                         "original_timezone",
		AlwaysOpenExternalLinks:          false,
		OpenExternalLinksInNewTab:        false,
		DisableBulkOperationsConfirmations: false,
	}

	// Test patching DisableBulkOperationsConfirmations
	trueValue := true
	req := &UserModificationRequest{
		DisableBulkOperationsConfirmations: &trueValue,
	}
	req.Patch(user)

	if !user.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be true after patch")
	}

	// Test that nil values don't modify the user
	user2 := &User{
		Username:                         "original",
		DisableBulkOperationsConfirmations: true,
	}
	req2 := &UserModificationRequest{} // All fields nil
	req2.Patch(user2)

	if user2.Username != "original" {
		t.Error("Username should not change when not provided in request")
	}
	if !user2.DisableBulkOperationsConfirmations {
		t.Error("DisableBulkOperationsConfirmations should remain true when not provided in request")
	}

	// Test patching false value
	falseValue := false
	user3 := &User{
		DisableBulkOperationsConfirmations: true,
	}
	req3 := &UserModificationRequest{
		DisableBulkOperationsConfirmations: &falseValue,
	}
	req3.Patch(user3)

	if user3.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be false after patch")
	}
}

func TestUserModificationRequestPatchMultipleFields(t *testing.T) {
	user := &User{
		Username:                         "original_user",
		Theme:                            "original_theme",
		AlwaysOpenExternalLinks:          false,
		OpenExternalLinksInNewTab:        false,
		DisableBulkOperationsConfirmations: false,
	}

	newUsername := "new_user"
	newTheme := "new_theme"
	trueValue := true

	req := &UserModificationRequest{
		Username:                         &newUsername,
		Theme:                            &newTheme,
		AlwaysOpenExternalLinks:          &trueValue,
		OpenExternalLinksInNewTab:        &trueValue,
		DisableBulkOperationsConfirmations: &trueValue,
	}
	req.Patch(user)

	if user.Username != "new_user" {
		t.Errorf("Expected Username to be 'new_user', got '%s'", user.Username)
	}
	if user.Theme != "new_theme" {
		t.Errorf("Expected Theme to be 'new_theme', got '%s'", user.Theme)
	}
	if !user.AlwaysOpenExternalLinks {
		t.Error("Expected AlwaysOpenExternalLinks to be true")
	}
	if !user.OpenExternalLinksInNewTab {
		t.Error("Expected OpenExternalLinksInNewTab to be true")
	}
	if !user.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be true")
	}
}
