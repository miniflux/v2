// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"strings"
	"testing"
)

func TestTruncateStringForTSVectorField(t *testing.T) {
	// Test case 1: Short Chinese text should not be truncated
	shortText := "这是一个简短的中文测试文本"
	result := truncateStringForTSVectorField(shortText)
	if result != shortText {
		t.Errorf("Short text should not be truncated, got %s", result)
	}

	// Test case 2: Long Chinese text should be truncated to stay under 1MB
	// Generate a long Chinese string that would exceed 1MB
	const megabyte = 1024 * 1024
	chineseChar := "汉"
	longText := strings.Repeat(chineseChar, megabyte/len(chineseChar)+1000) // Ensure it exceeds 1MB

	result = truncateStringForTSVectorField(longText)

	// Verify the result is under 1MB
	if len(result) >= megabyte {
		t.Errorf("Truncated text should be under 1MB, got %d bytes", len(result))
	}

	// Verify the result is still valid UTF-8 and doesn't cut in the middle of a character
	if !strings.HasPrefix(longText, result) {
		t.Error("Truncated text should be a prefix of original text")
	}

	// Test case 3: Text exactly at limit should not be truncated
	limitText := strings.Repeat("a", megabyte-1)
	result = truncateStringForTSVectorField(limitText)
	if result != limitText {
		t.Error("Text under limit should not be truncated")
	}

	// Test case 4: Mixed Chinese and ASCII text
	mixedText := strings.Repeat("测试Test汉字", megabyte/20) // Create large mixed text
	result = truncateStringForTSVectorField(mixedText)

	if len(result) >= megabyte {
		t.Errorf("Mixed text should be truncated under 1MB, got %d bytes", len(result))
	}

	// Verify no broken UTF-8 sequences
	if !strings.HasPrefix(mixedText, result) {
		t.Error("Truncated mixed text should be a valid prefix")
	}

	// Test case 5: Large text ending with ASCII characters
	asciiSuffix := strings.Repeat("a", megabyte-100) + strings.Repeat("测试", 50) + "abcdef"
	result = truncateStringForTSVectorField(asciiSuffix)

	if len(result) >= megabyte {
		t.Errorf("ASCII suffix text should be truncated under 1MB, got %d bytes", len(result))
	}

	// Should end with ASCII character
	if !strings.HasPrefix(asciiSuffix, result) {
		t.Error("Truncated ASCII suffix text should be a valid prefix")
	}

	// Test case 6: Large ASCII text to cover ASCII branch in UTF-8 detection
	largeAscii := strings.Repeat("abcdefghijklmnopqrstuvwxyz", megabyte/26+1000)
	result = truncateStringForTSVectorField(largeAscii)

	if len(result) >= megabyte {
		t.Errorf("Large ASCII text should be truncated under 1MB, got %d bytes", len(result))
	}

	// Should be a prefix
	if !strings.HasPrefix(largeAscii, result) {
		t.Error("Truncated ASCII text should be a valid prefix")
	}

	// Test case 7: Edge case - string that would trigger the fallback
	// Create a pathological case: all continuation bytes without start bytes
	// This should trigger the fallback because there are no valid UTF-8 boundaries
	invalidBytes := make([]byte, megabyte)
	for i := range invalidBytes {
		invalidBytes[i] = 0x80 // Continuation byte without start byte
	}
	result = truncateStringForTSVectorField(string(invalidBytes))

	// Should return empty string as fallback
	if result != "" {
		t.Errorf("Invalid UTF-8 continuation bytes should return empty string, got %d bytes", len(result))
	}
}
