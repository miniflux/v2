// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package media // import "miniflux.app/v2/internal/reader/media"

import "testing"

func TestContentMimeType(t *testing.T) {
	scenarios := []struct {
		inputType, inputMedium, expectedMimeType string
	}{
		{"image/png", "image", "image/png"},
		{"", "image", "image/*"},
		{"", "video", "video/*"},
		{"", "audio", "audio/*"},
		{"", "", "application/octet-stream"},
	}

	for _, scenario := range scenarios {
		content := &Content{Type: scenario.inputType, Medium: scenario.inputMedium}
		result := content.MimeType()
		if result != scenario.expectedMimeType {
			t.Errorf(`Unexpected mime type, got %q instead of %q for type=%q medium=%q`,
				result,
				scenario.expectedMimeType,
				scenario.inputType,
				scenario.inputMedium,
			)
		}
	}
}

func TestContentSize(t *testing.T) {
	scenarios := []struct {
		inputSize    string
		expectedSize int64
	}{
		{"", 0},
		{"123", int64(123)},
	}

	for _, scenario := range scenarios {
		content := &Content{FileSize: scenario.inputSize}
		result := content.Size()
		if result != scenario.expectedSize {
			t.Errorf(`Unexpected size, got %d instead of %d for %q`,
				result,
				scenario.expectedSize,
				scenario.inputSize,
			)
		}
	}
}

func TestPeerLinkType(t *testing.T) {
	scenarios := []struct {
		inputType        string
		expectedMimeType string
	}{
		{"", "application/octet-stream"},
		{"application/x-bittorrent", "application/x-bittorrent"},
	}

	for _, scenario := range scenarios {
		peerLink := &PeerLink{Type: scenario.inputType}
		result := peerLink.MimeType()
		if result != scenario.expectedMimeType {
			t.Errorf(`Unexpected mime type, got %q instead of %q for %q`,
				result,
				scenario.expectedMimeType,
				scenario.inputType,
			)
		}
	}
}

func TestDescription(t *testing.T) {
	scenarios := []struct {
		inputType           string
		inputContent        string
		expectedDescription string
	}{
		{"", "", ""},
		{"html", "a <b>c</b>", "a <b>c</b>"},
		{"plain", "a\nhttp://www.example.org/", `a<br><a href="http://www.example.org/">http://www.example.org/</a>`},
	}

	for _, scenario := range scenarios {
		desc := &Description{Type: scenario.inputType, Description: scenario.inputContent}
		result := desc.HTML()
		if result != scenario.expectedDescription {
			t.Errorf(`Unexpected description, got %q instead of %q for %q`,
				result,
				scenario.expectedDescription,
				scenario.inputType,
			)
		}
	}
}

func TestFirstDescription(t *testing.T) {
	var descList DescriptionList
	descList = append(descList, Description{})
	descList = append(descList, Description{Description: "Something"})

	if descList.First() != "Something" {
		t.Errorf(`Unexpected description`)
	}
}
