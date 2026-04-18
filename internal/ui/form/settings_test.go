// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNewSettingsFormDisableBulkOperationsConfirmations(t *testing.T) {
	// Test when checkbox is checked (value="1")
	formData := url.Values{}
	formData.Set("disable_bulk_operations_confirmations", "1")
	req := &http.Request{
		Method: "POST",
		PostForm: formData,
	}

	form := NewSettingsForm(req)
	if !form.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be true when form value is '1'")
	}

	// Test when checkbox is unchecked (no value)
	formData2 := url.Values{}
	req2 := &http.Request{
		Method:   "POST",
		PostForm: formData2,
	}

	form2 := NewSettingsForm(req2)
	if form2.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be false when form value is not provided")
	}
}

func TestSettingsFormMergeDisableBulkOperationsConfirmations(t *testing.T) {
	// Skip this test as Merge() depends on global config.Opts which requires
	// proper initialization. The Merge function is tested through integration tests.
	t.Skip("Skipping test: Merge() depends on global config.Opts")
}

func TestSettingsFormDisableBulkOperationsConfirmationsFromRequestBody(t *testing.T) {
	// Test parsing from request body (for PUT requests)
	body := strings.NewReader("disable_bulk_operations_confirmations=1")
	req, err := http.NewRequest("PUT", "/settings", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	form := NewSettingsForm(req)
	if !form.DisableBulkOperationsConfirmations {
		t.Error("Expected DisableBulkOperationsConfirmations to be true when parsed from request body")
	}
}
