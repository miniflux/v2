// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"miniflux.app/v2/internal/config"
)

func TestEnclosure_Html5MimeTypeGivesOriginalMimeType(t *testing.T) {
	enclosure := Enclosure{MimeType: "thing/thisMimeTypeIsNotExpectedToBeReplaced"}
	if enclosure.Html5MimeType() != enclosure.MimeType {
		t.Fatalf(
			"HTML5 MimeType must provide original MimeType if not explicitly Replaced. Got %s ,expected '%s' ",
			enclosure.Html5MimeType(),
			enclosure.MimeType,
		)
	}
}

func TestEnclosure_Html5MimeTypeReplaceStandardM4vByAppleSpecificMimeType(t *testing.T) {
	enclosure := Enclosure{MimeType: "video/m4v"}
	if enclosure.Html5MimeType() != "video/x-m4v" {
		// Solution from this stackoverflow discussion:
		// https://stackoverflow.com/questions/15277147/m4v-mimetype-video-mp4-or-video-m4v/66945470#66945470
		// tested at the time of this commit (06/2023) on latest Firefox & Vivaldi on this feed
		// https://www.florenceporcel.com/podcast/lfhdu.xml
		t.Fatalf(
			"HTML5 MimeType must be replaced by 'video/x-m4v' when originally video/m4v to ensure playbacks in browsers. Got '%s'",
			enclosure.Html5MimeType(),
		)
	}
}

func TestEnclosure_IsAudio(t *testing.T) {
	testCases := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"MP3 audio", "audio/mpeg", true},
		{"WAV audio", "audio/wav", true},
		{"OGG audio", "audio/ogg", true},
		{"Mixed case audio", "Audio/MP3", true},
		{"Video file", "video/mp4", false},
		{"Image file", "image/jpeg", false},
		{"Text file", "text/plain", false},
		{"Empty mime type", "", false},
		{"Audio with extra info", "audio/mpeg; charset=utf-8", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enclosure := &Enclosure{MimeType: tc.mimeType}
			if got := enclosure.IsAudio(); got != tc.expected {
				t.Errorf("IsAudio() = %v, want %v for mime type %s", got, tc.expected, tc.mimeType)
			}
		})
	}
}

func TestEnclosure_IsVideo(t *testing.T) {
	testCases := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"MP4 video", "video/mp4", true},
		{"AVI video", "video/avi", true},
		{"WebM video", "video/webm", true},
		{"M4V video", "video/m4v", true},
		{"Mixed case video", "Video/MP4", true},
		{"Audio file", "audio/mpeg", false},
		{"Image file", "image/jpeg", false},
		{"Text file", "text/plain", false},
		{"Empty mime type", "", false},
		{"Video with extra info", "video/mp4; codecs=\"avc1.42E01E\"", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enclosure := &Enclosure{MimeType: tc.mimeType}
			if got := enclosure.IsVideo(); got != tc.expected {
				t.Errorf("IsVideo() = %v, want %v for mime type %s", got, tc.expected, tc.mimeType)
			}
		})
	}
}

func TestEnclosure_IsImage(t *testing.T) {
	testCases := []struct {
		name     string
		mimeType string
		url      string
		expected bool
	}{
		{"JPEG image by mime", "image/jpeg", "http://example.com/file", true},
		{"PNG image by mime", "image/png", "http://example.com/file", true},
		{"GIF image by mime", "image/gif", "http://example.com/file", true},
		{"Mixed case image mime", "Image/JPEG", "http://example.com/file", true},
		{"JPG file extension", "application/octet-stream", "http://example.com/photo.jpg", true},
		{"JPEG file extension", "text/plain", "http://example.com/photo.jpeg", true},
		{"PNG file extension", "unknown/type", "http://example.com/photo.png", true},
		{"GIF file extension", "binary/data", "http://example.com/photo.gif", true},
		{"Mixed case extension", "text/plain", "http://example.com/photo.JPG", true},
		{"Image mime and extension", "image/jpeg", "http://example.com/photo.jpg", true},
		{"Video file", "video/mp4", "http://example.com/video.mp4", false},
		{"Audio file", "audio/mpeg", "http://example.com/audio.mp3", false},
		{"Text file", "text/plain", "http://example.com/file.txt", false},
		{"No extension", "text/plain", "http://example.com/file", false},
		{"Other extension", "text/plain", "http://example.com/file.pdf", false},
		{"Empty values", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enclosure := &Enclosure{MimeType: tc.mimeType, URL: tc.url}
			if got := enclosure.IsImage(); got != tc.expected {
				t.Errorf("IsImage() = %v, want %v for mime type %s and URL %s", got, tc.expected, tc.mimeType, tc.url)
			}
		})
	}
}

func TestEnclosureList_FindMediaPlayerEnclosure(t *testing.T) {
	testCases := []struct {
		name        string
		enclosures  EnclosureList
		expectedNil bool
	}{
		{
			name: "Returns first audio enclosure",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
				&Enclosure{URL: "http://example.com/video.mp4", MimeType: "video/mp4"},
			},
			expectedNil: false,
		},
		{
			name: "Returns first video enclosure",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/video.mp4", MimeType: "video/mp4"},
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
			},
			expectedNil: false,
		},
		{
			name: "Skips image enclosure and returns audio",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/image.jpg", MimeType: "image/jpeg"},
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
			},
			expectedNil: false,
		},
		{
			name: "Skips enclosure with empty URL",
			enclosures: EnclosureList{
				&Enclosure{URL: "", MimeType: "audio/mpeg"},
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
			},
			expectedNil: false,
		},
		{
			name: "Returns nil for no media enclosures",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/image.jpg", MimeType: "image/jpeg"},
				&Enclosure{URL: "http://example.com/doc.pdf", MimeType: "application/pdf"},
			},
			expectedNil: true,
		},
		{
			name:        "Returns nil for empty list",
			enclosures:  EnclosureList{},
			expectedNil: true,
		},
		{
			name: "Returns nil for all empty URLs",
			enclosures: EnclosureList{
				&Enclosure{URL: "", MimeType: "audio/mpeg"},
				&Enclosure{URL: "", MimeType: "video/mp4"},
			},
			expectedNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.enclosures.FindMediaPlayerEnclosure()
			if tc.expectedNil {
				if result != nil {
					t.Errorf("FindMediaPlayerEnclosure() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("FindMediaPlayerEnclosure() = nil, want non-nil")
				} else if !result.IsAudio() && !result.IsVideo() {
					t.Errorf("FindMediaPlayerEnclosure() returned non-media enclosure: %s", result.MimeType)
				}
			}
		})
	}
}

func TestEnclosureList_ContainsAudioOrVideo(t *testing.T) {
	testCases := []struct {
		name       string
		enclosures EnclosureList
		expected   bool
	}{
		{
			name: "Contains audio",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "audio/mpeg"},
				&Enclosure{MimeType: "image/jpeg"},
			},
			expected: true,
		},
		{
			name: "Contains video",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "image/jpeg"},
				&Enclosure{MimeType: "video/mp4"},
			},
			expected: true,
		},
		{
			name: "Contains both audio and video",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "audio/mpeg"},
				&Enclosure{MimeType: "video/mp4"},
			},
			expected: true,
		},
		{
			name: "Contains only images",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "image/jpeg"},
				&Enclosure{MimeType: "image/png"},
			},
			expected: false,
		},
		{
			name: "Contains only documents",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "application/pdf"},
				&Enclosure{MimeType: "text/plain"},
			},
			expected: false,
		},
		{
			name:       "Empty list",
			enclosures: EnclosureList{},
			expected:   false,
		},
		{
			name: "Single audio enclosure",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "audio/wav"},
			},
			expected: true,
		},
		{
			name: "Single video enclosure",
			enclosures: EnclosureList{
				&Enclosure{MimeType: "video/webm"},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.enclosures.ContainsAudioOrVideo()
			if result != tc.expected {
				t.Errorf("ContainsAudioOrVideo() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestEnclosure_ProxifyEnclosureURL(t *testing.T) {
	// Initialize config for testing
	os.Clearenv()
	os.Setenv("BASE_URL", "http://localhost")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test-private-key")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	testCases := []struct {
		name                    string
		url                     string
		mimeType                string
		mediaProxyOption        string
		mediaProxyResourceTypes []string
		expectedURLChanged      bool
	}{
		{
			name:                    "HTTP URL with audio type - proxy mode all",
			url:                     "http://example.com/audio.mp3",
			mimeType:                "audio/mpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      true,
		},
		{
			name:                    "HTTPS URL with video type - proxy mode all",
			url:                     "https://example.com/video.mp4",
			mimeType:                "video/mp4",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      true,
		},
		{
			name:                    "HTTP URL with video type - proxy mode http-only",
			url:                     "http://example.com/video.mp4",
			mimeType:                "video/mp4",
			mediaProxyOption:        "http-only",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      true,
		},
		{
			name:                    "HTTPS URL with video type - proxy mode http-only",
			url:                     "https://example.com/video.mp4",
			mimeType:                "video/mp4",
			mediaProxyOption:        "http-only",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      false,
		},
		{
			name:                    "HTTP URL with image type - not in resource types",
			url:                     "http://example.com/image.jpg",
			mimeType:                "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      false,
		},
		{
			name:                    "HTTP URL with image type - in resource types",
			url:                     "http://example.com/image.jpg",
			mimeType:                "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video", "image"},
			expectedURLChanged:      true,
		},
		{
			name:                    "HTTP URL - proxy mode none",
			url:                     "http://example.com/audio.mp3",
			mimeType:                "audio/mpeg",
			mediaProxyOption:        "none",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      false,
		},
		{
			name:                    "Empty URL",
			url:                     "",
			mimeType:                "audio/mpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      false,
		},
		{
			name:                    "Non-media MIME type",
			url:                     "http://example.com/doc.pdf",
			mimeType:                "application/pdf",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedURLChanged:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enclosure := &Enclosure{
				URL:      tc.url,
				MimeType: tc.mimeType,
			}

			originalURL := enclosure.URL

			// Call the method
			enclosure.ProxifyEnclosureURL(router, tc.mediaProxyOption, tc.mediaProxyResourceTypes)

			// Check if URL changed as expected
			urlChanged := enclosure.URL != originalURL
			if urlChanged != tc.expectedURLChanged {
				t.Errorf("ProxifyEnclosureURL() URL changed = %v, want %v. Original: %s, New: %s",
					urlChanged, tc.expectedURLChanged, originalURL, enclosure.URL)
			}

			// If URL should have changed, verify it's not empty
			if tc.expectedURLChanged && enclosure.URL == "" {
				t.Error("ProxifyEnclosureURL() resulted in empty URL when proxification was expected")
			}

			// If URL shouldn't have changed, verify it's identical
			if !tc.expectedURLChanged && enclosure.URL != originalURL {
				t.Errorf("ProxifyEnclosureURL() URL changed unexpectedly from %s to %s", originalURL, enclosure.URL)
			}
		})
	}
}

func TestEnclosureList_ProxifyEnclosureURL(t *testing.T) {
	// Initialize config for testing
	os.Clearenv()
	os.Setenv("BASE_URL", "http://localhost")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test-private-key")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	testCases := []struct {
		name                    string
		enclosures              EnclosureList
		mediaProxyOption        string
		mediaProxyResourceTypes []string
		expectedChangedCount    int
	}{
		{
			name: "Mixed enclosures with all proxy mode",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
				&Enclosure{URL: "https://example.com/video.mp4", MimeType: "video/mp4"},
				&Enclosure{URL: "http://example.com/image.jpg", MimeType: "image/jpeg"},
				&Enclosure{URL: "http://example.com/doc.pdf", MimeType: "application/pdf"},
			},
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedChangedCount:    2, // audio and video should be proxified
		},
		{
			name: "Mixed enclosures with http-only proxy mode",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
				&Enclosure{URL: "https://example.com/video.mp4", MimeType: "video/mp4"},
				&Enclosure{URL: "http://example.com/video2.mp4", MimeType: "video/mp4"},
			},
			mediaProxyOption:        "http-only",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedChangedCount:    2, // only HTTP URLs should be proxified
		},
		{
			name: "No media types in resource list",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
				&Enclosure{URL: "http://example.com/video.mp4", MimeType: "video/mp4"},
			},
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expectedChangedCount:    0, // no matching resource types
		},
		{
			name: "Proxy mode none",
			enclosures: EnclosureList{
				&Enclosure{URL: "http://example.com/audio.mp3", MimeType: "audio/mpeg"},
				&Enclosure{URL: "http://example.com/video.mp4", MimeType: "video/mp4"},
			},
			mediaProxyOption:        "none",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedChangedCount:    0,
		},
		{
			name:                    "Empty enclosure list",
			enclosures:              EnclosureList{},
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedChangedCount:    0,
		},
		{
			name: "Enclosures with empty URLs",
			enclosures: EnclosureList{
				&Enclosure{URL: "", MimeType: "audio/mpeg"},
				&Enclosure{URL: "http://example.com/video.mp4", MimeType: "video/mp4"},
			},
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio", "video"},
			expectedChangedCount:    1, // only the non-empty URL should be processed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Store original URLs
			originalURLs := make([]string, len(tc.enclosures))
			for i, enclosure := range tc.enclosures {
				originalURLs[i] = enclosure.URL
			}

			// Call the method
			tc.enclosures.ProxifyEnclosureURL(router, tc.mediaProxyOption, tc.mediaProxyResourceTypes)

			// Count how many URLs actually changed
			changedCount := 0
			for i, enclosure := range tc.enclosures {
				if enclosure.URL != originalURLs[i] {
					changedCount++
					// Verify that changed URLs are not empty (unless they were empty originally)
					if originalURLs[i] != "" && enclosure.URL == "" {
						t.Errorf("Enclosure %d: ProxifyEnclosureURL resulted in empty URL", i)
					}
				}
			}

			if changedCount != tc.expectedChangedCount {
				t.Errorf("ProxifyEnclosureURL() changed %d URLs, want %d", changedCount, tc.expectedChangedCount)
			}
		})
	}
}

func TestEnclosure_ProxifyEnclosureURL_EdgeCases(t *testing.T) {
	// Initialize config for testing
	os.Clearenv()
	os.Setenv("BASE_URL", "http://localhost")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test-private-key")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")
	t.Run("Empty resource types slice", func(t *testing.T) {
		enclosure := &Enclosure{
			URL:      "http://example.com/audio.mp3",
			MimeType: "audio/mpeg",
		}

		originalURL := enclosure.URL
		enclosure.ProxifyEnclosureURL(router, "all", []string{})

		// With empty resource types, URL should not change
		if enclosure.URL != originalURL {
			t.Errorf("URL should not change with empty resource types. Original: %s, New: %s", originalURL, enclosure.URL)
		}
	})

	t.Run("Nil resource types slice", func(t *testing.T) {
		enclosure := &Enclosure{
			URL:      "http://example.com/audio.mp3",
			MimeType: "audio/mpeg",
		}

		originalURL := enclosure.URL
		enclosure.ProxifyEnclosureURL(router, "all", nil)

		// With nil resource types, URL should not change
		if enclosure.URL != originalURL {
			t.Errorf("URL should not change with nil resource types. Original: %s, New: %s", originalURL, enclosure.URL)
		}
	})
	t.Run("Invalid proxy mode", func(t *testing.T) {
		enclosure := &Enclosure{
			URL:      "http://example.com/audio.mp3",
			MimeType: "audio/mpeg",
		}

		originalURL := enclosure.URL
		enclosure.ProxifyEnclosureURL(router, "invalid-mode", []string{"audio"})

		// With invalid proxy mode, the function still proxifies non-HTTPS URLs
		// because shouldProxifyURL defaults to checking URL scheme
		if enclosure.URL == originalURL {
			t.Errorf("URL should change for HTTP URL even with invalid proxy mode. Original: %s, New: %s", originalURL, enclosure.URL)
		}
	})
}
