// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package urllib // import "miniflux.app/v2/internal/urllib"

import (
	"net"
	"net/url"
	"testing"
)

func TestIsRelativePath(t *testing.T) {
	scenarios := map[string]bool{
		// Valid relative paths
		"path/to/file.ext":    true,
		"./path/to/file.ext":  true,
		"../path/to/file.ext": true,
		"file.ext":            true,
		"./file.ext":          true,
		"../file.ext":         true,
		"/absolute/path":      true,
		"path?query=value":    true,
		"path#fragment":       true,
		"path?query#fragment": true,

		// Not relative paths
		"https://example.org/file.ext": false,
		"http://example.org/file.ext":  false,
		"//example.org/file.ext":       false,
		"//example.org":                false,
		"ftp://example.org/file.ext":   false,
		"mailto:user@example.org":      false,
		"magnet:?xt=urn:btih:example":  false,
		"":                             false,
		"magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C": false,
	}

	for input, expected := range scenarios {
		actual := IsRelativePath(input)
		if actual != expected {
			t.Errorf(`Unexpected result for IsRelativePath, got %v instead of %v for %q`, actual, expected, input)
		}
	}
}

func TestIsAbsoluteURL(t *testing.T) {
	scenarios := map[string]bool{
		"https://example.org/file.pdf": true,
		"magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7": true,
		"invalid url":    false,
		"/relative/path": false,
	}

	for input, expected := range scenarios {
		actual := IsAbsoluteURL(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v for %q`, actual, expected, input)
		}
	}
}

func TestAbsoluteURL(t *testing.T) {
	type absoluteScenario struct {
		name          string
		base          string
		relative      string
		expected      string
		wantErr       bool
		runWithParsed bool
		useNilParsed  bool
	}

	scenarios := []absoluteScenario{
		{"absolute path", "https://example.org/folder/", "/path/file.ext", "https://example.org/path/file.ext", false, true, false},
		{"relative path", "https://example.org/folder/", "path/file.ext", "https://example.org/folder/path/file.ext", false, true, false},
		{"dot path root", "https://example.org/path", "./", "https://example.org/", false, true, false},
		{"dot path folder", "https://example.org/folder/", "./", "https://example.org/folder/", false, true, false},
		{"missing slash in base", "https://example.org/folder", "path/file.ext", "https://example.org/path/file.ext", false, true, false},
		{"already absolute", "https://example.org/folder/", "https://example.org/path/file.ext", "https://example.org/path/file.ext", false, true, false},
		{"protocol relative", "https://www.example.org/", "//static.example.org/path/file.ext", "https://static.example.org/path/file.ext", false, true, false},
		{"magnet keeps scheme", "https://www.example.org/", "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a", "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a", false, true, false},
		{"magnet with query", "https://www.example.org/", "magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7", "magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7", false, true, false},
		{"empty relative returns base", "https://example.org/folder/", "", "https://example.org/folder/", false, true, false},
		{"invalid base errors", "://bad", "path/file.ext", "", true, false, false},
		{"absolute ignores invalid base", "://bad", "https://example.org/path/file.ext", "https://example.org/path/file.ext", false, true, true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			actual, err := ResolveToAbsoluteURL(scenario.base, scenario.relative)
			if scenario.wantErr {
				if err == nil {
					t.Fatalf("expected error for base %q relative %q", scenario.base, scenario.relative)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for base %q relative %q: %v", scenario.base, scenario.relative, err)
			}
			if actual != scenario.expected {
				t.Fatalf("unexpected result, got %q instead of %q for (%q, %q)", actual, scenario.expected, scenario.base, scenario.relative)
			}

			if scenario.runWithParsed {
				var parsedBase *url.URL
				if !scenario.useNilParsed && scenario.base != "" {
					var parseErr error
					parsedBase, parseErr = url.Parse(scenario.base)
					if parseErr != nil {
						t.Fatalf("unable to parse base %q: %v", scenario.base, parseErr)
					}
				}

				actualParsed, errParsed := ResolveToAbsoluteURLWithParsedBaseURL(parsedBase, scenario.relative)
				if errParsed != nil {
					t.Fatalf("unexpected error with parsed base for (%q, %q): %v", scenario.base, scenario.relative, errParsed)
				}
				if actualParsed != scenario.expected {
					t.Fatalf("unexpected parsed-base result, got %q instead of %q for (%q, %q)", actualParsed, scenario.expected, scenario.base, scenario.relative)
				}
			}
		})
	}
}

func TestRootURL(t *testing.T) {
	scenarios := map[string]string{
		"":                                  "",
		"https://example.org/path/file.ext": "https://example.org/",
		"https://example.org/path/file.ext?test=abc": "https://example.org/",
		"//static.example.org/path/file.ext":         "https://static.example.org/",
		"https://example|org/path/file.ext":          "https://example|org/path/file.ext",
		"/relative/path":                             "/relative/path",
		"http://example.org:8080/path":               "http://example.org:8080/",
	}

	for input, expected := range scenarios {
		actual := RootURL(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, actual, expected)
		}
	}
}

func TestIsHTTPS(t *testing.T) {
	scenarios := map[string]bool{
		"https://example.org/": true,
		"http://example.org/":  false,
		"https://example|org/": false,
	}

	for input, expected := range scenarios {
		actual := IsHTTPS(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v`, actual, expected)
		}
	}
}

func TestDomain(t *testing.T) {
	scenarios := map[string]string{
		"https://static.example.org/": "static.example.org",
		"https://example|org/":        "https://example|org/",
	}

	for input, expected := range scenarios {
		actual := Domain(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, actual, expected)
		}
	}
}

func TestDomainWithoutWWW(t *testing.T) {
	scenarios := map[string]string{
		"https://www.example.org/":     "example.org",
		"https://example.org/":         "example.org",
		"https://www.sub.example.org/": "sub.example.org",
		"https://example|org/":         "https://example|org/",
	}

	for input, expected := range scenarios {
		actual := DomainWithoutWWW(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, actual, expected)
		}
	}
}

func TestJoinBaseURLAndPath(t *testing.T) {
	type args struct {
		baseURL string
		path    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty base url", args{"", "/api/bookmarks/"}, "", true},
		{"empty path", args{"https://example.com", ""}, "", true},
		{"invalid base url", args{"incorrect url", ""}, "", true},
		{"valid", args{"https://example.com", "/api/bookmarks/"}, "https://example.com/api/bookmarks/", false},
		{"valid", args{"https://example.com/subfolder", "/api/bookmarks/"}, "https://example.com/subfolder/api/bookmarks/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JoinBaseURLAndPath(tt.args.baseURL, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinBaseURLAndPath error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JoinBaseURLAndPath = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNonPublicIP(t *testing.T) {
	testCases := []struct {
		name     string
		ipString string
		want     bool
	}{
		{"nil", "", true},
		{"private IPv4", "192.168.1.10", true},
		{"loopback IPv4", "127.0.0.1", true},
		{"link-local IPv4", "169.254.42.1", true},
		{"multicast IPv4", "224.0.0.1", true},
		{"unspecified IPv6", "::", true},
		{"loopback IPv6", "::1", true},
		{"multicast IPv6", "ff02::1", true},
		{"public IPv4", "93.184.216.34", false},
		{"public IPv6", "2001:4860:4860::8888", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ip net.IP
			if tc.ipString != "" {
				ip = net.ParseIP(tc.ipString)
				if ip == nil {
					t.Fatalf("unable to parse %q", tc.ipString)
				}
			}

			if got := isNonPublicIP(ip); got != tc.want {
				t.Fatalf("unexpected result for %s: got %v want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestResolvesToPrivateIP(t *testing.T) {
	testCases := []struct {
		name string
		host string
		want bool
	}{
		{"localhost", "localhost", true},
		{"example.org", "example.org", false},
		{"loopback IPv4 literal", "127.0.0.1", true},
		{"loopback IPv6 literal", "::1", true},
		{"private IPv4 literal", "192.168.1.1", true},
		{"public IPv4 literal", "93.184.216.34", false},
		{"public IPv6 literal", "2001:4860:4860::8888", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolvesToPrivateIP(tc.host)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.host, err)
			}
			if got != tc.want {
				t.Fatalf("unexpected result for %s: got %v want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestResolvesToPrivateIPError(t *testing.T) {
	if _, err := ResolvesToPrivateIP(""); err == nil {
		t.Fatalf("expected an error for empty host")
	}
}
