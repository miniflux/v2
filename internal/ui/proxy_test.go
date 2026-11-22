// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/mediaproxy"
)

func TestIsInlineMediaType(t *testing.T) {
	scenarios := []struct {
		contentType string
		expected    bool
	}{
		{"image/png", true},
		{"image/svg+xml", true},
		{"IMAGE/JPEG; charset=utf-8", true},
		{"audio/mpeg", true},
		{"video/mp4", true},
		{"text/html", false},
		{"text/html; charset=utf-8", false},
		{"application/xhtml+xml", false},
		{"application/octet-stream", false},
		{"text/plain", false},
		{"", false},
		{";;", false},
		{"image", false},
	}

	for _, scenario := range scenarios {
		if result := isInlineMediaType(scenario.contentType); result != scenario.expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, scenario.contentType, result, scenario.expected)
		}
	}
}

func TestMediaProxyContentTypeHandling(t *testing.T) {
	scenarios := []struct {
		name                       string
		upstreamPath               string
		upstreamContentType        string
		expectedContentType        string
		expectedContentDisposition string
	}{
		{"png image", "/image.png", "image/png", "image/png", "inline; filename=image.png"},
		{"svg image with parameters", "/image.svg", "IMAGE/SVG+XML; charset=utf-8", "IMAGE/SVG+XML; charset=utf-8", "inline; filename=image.svg"},
		{"audio", "/episode.mp3", "audio/mpeg", "audio/mpeg", "inline; filename=episode.mp3"},
		{"video", "/clip.mp4", "video/mp4", "video/mp4", "inline; filename=clip.mp4"},
		{"image without filename", "/", "image/png", "image/png", ""},
		{"html page", "/page.html", "text/html; charset=utf-8", "application/octet-stream", "attachment; filename=page.html"},
		{"xhtml page", "/page.xhtml", "application/xhtml+xml", "application/octet-stream", "attachment; filename=page.xhtml"},
		{"html page without filename", "/", "text/html", "application/octet-stream", "attachment"},
		{"missing content type", "/unknown", "", "application/octet-stream", "attachment; filename=unknown"},
		{"malformed content type", "/broken", ";;", "application/octet-stream", "attachment; filename=broken"},
	}

	const upstreamBody = "<html><body>upstream body</body></html>"

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			configureTestOptions(t, map[string]string{
				"MEDIA_PROXY_PRIVATE_KEY":        "test-private-key",
				"FETCHER_ALLOW_PRIVATE_NETWORKS": "1",
			})

			upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if scenario.upstreamContentType == "" {
					// Setting the header to nil prevents net/http from sniffing a Content-Type.
					w.Header()["Content-Type"] = nil
				} else {
					w.Header().Set("Content-Type", scenario.upstreamContentType)
				}
				io.WriteString(w, upstreamBody)
			}))
			t.Cleanup(upstreamServer.Close)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /proxy/{encodedDigest}/{encodedURL}", (&handler{}).mediaProxy)

			proxyPath := mediaproxy.ProxifyRelativeURL(upstreamServer.URL + scenario.upstreamPath)
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, proxyPath, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf(`Unexpected status code, got %d instead of %d`, recorder.Code, http.StatusOK)
			}

			if contentType := recorder.Header().Get("Content-Type"); contentType != scenario.expectedContentType {
				t.Errorf(`Unexpected Content-Type, got %q instead of %q`, contentType, scenario.expectedContentType)
			}

			if contentDisposition := recorder.Header().Get("Content-Disposition"); contentDisposition != scenario.expectedContentDisposition {
				t.Errorf(`Unexpected Content-Disposition, got %q instead of %q`, contentDisposition, scenario.expectedContentDisposition)
			}

			if csp := recorder.Header().Get("Content-Security-Policy"); csp != response.ContentSecurityPolicyForUntrustedContent {
				t.Errorf(`Unexpected Content-Security-Policy, got %q`, csp)
			}

			if body := recorder.Body.String(); body != upstreamBody {
				t.Errorf(`Unexpected body, got %q instead of %q`, body, upstreamBody)
			}
		})
	}
}
