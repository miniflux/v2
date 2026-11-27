// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestConvertEntryIDToLongFormItemID(t *testing.T) {
	entryID := int64(344691561)
	expected := "tag:google.com,2005:reader/item/00000000148b9369"
	result := convertEntryIDToLongFormItemID(entryID)

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseItemIDsFromRequest(t *testing.T) {
	formValues := url.Values{}
	formValues.Add("i", "12345")
	formValues.Add("i", "tag:google.com,2005:reader/item/00000000148b9369")
	formValues.Add("i", "tag:google.com,2005:reader/item/2f2")
	formValues.Add("i", "000000000000046f")
	formValues.Add("i", "tag:google.com,2005:reader/item/272")

	request := &http.Request{
		Form: formValues,
	}

	result, err := parseItemIDsFromRequest(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var expected = []int64{12345, 344691561, 754, 1135, 626}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test with no item IDs
	formValues = url.Values{}
	request = &http.Request{
		Form: formValues,
	}
	_, err = parseItemIDsFromRequest(request)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestParseItemID(t *testing.T) {
	// Test with long form ID and hex ID
	result, err := parseItemID("tag:google.com,2005:reader/item/0000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int64(1)
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}

	// Test with hexadecimal long form ID
	result, err = parseItemID("0000000000000468")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected = int64(1128)
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}

	// Test with short form ID
	result, err = parseItemID("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected = int64(12345)
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}

	// Test with invalid long form ID
	_, err = parseItemID("tag:google.com,2005:reader/item/000000000000000g")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Test with invalid short form ID
	_, err = parseItemID("invalid_id")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Test with empty ID
	_, err = parseItemID("")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestParseItemIDParsesEveryDigit(t *testing.T) {
	result, err := parseItemID("tag:google.com,2005:reader/item/00000000000000010")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 16 {
		t.Errorf("expected 16, got %d", result)
	}
}

func TestParseItemIDRejectsMalformedIDs(t *testing.T) {
	for _, itemID := range []string{
		"tag:google.com,2005:reader/item/0000000000000001garbage",
		"tag:google.com,2005:reader/item/2f2xyz",
		"tag:google.com,2005:reader/item/ 2f2",
		"tag:google.com,2005:reader/item/+2f2",
		"tag:google.com,2005:reader/item/0x2f2",
		"tag:google.com,2005:reader/item/",
		"tag:google.com,2005:reader/item/8000000000000000",
		"000000000000048g",
		"00000000000001-5",
		"+12345",
		"12345garbage",
	} {
		t.Run(itemID, func(t *testing.T) {
			if result, err := parseItemID(itemID); err == nil {
				t.Fatalf("expected error for malformed item ID %q, got %d", itemID, result)
			}
		})
	}
}

func TestParseItemIDRejectsNonPositiveIDs(t *testing.T) {
	for _, itemID := range []string{
		"0",
		"-1",
		"0000000000000000",
		"tag:google.com,2005:reader/item/0000000000000000",
	} {
		t.Run(itemID, func(t *testing.T) {
			if _, err := parseItemID(itemID); err == nil {
				t.Fatalf("expected error for non-positive item ID %q", itemID)
			}
		})
	}
}
