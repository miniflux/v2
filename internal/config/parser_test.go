// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"reflect"
	"testing"
)

func TestParseBoolValue(t *testing.T) {
	scenarios := map[string]bool{
		"":        true,
		"1":       true,
		"Yes":     true,
		"yes":     true,
		"True":    true,
		"true":    true,
		"on":      true,
		"false":   false,
		"off":     false,
		"invalid": false,
	}

	for input, expected := range scenarios {
		result := parseBool(input, true)
		if result != expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, input, result, expected)
		}
	}
}

func TestParseStringValueWithUnsetVariable(t *testing.T) {
	if parseString("", "defaultValue") != "defaultValue" {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseStringValue(t *testing.T) {
	if parseString("test", "defaultValue") != "test" {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestParseIntValueWithUnsetVariable(t *testing.T) {
	if parseInt("", 42) != 42 {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseIntValueWithInvalidInput(t *testing.T) {
	if parseInt("invalid integer", 42) != 42 {
		t.Errorf(`Invalid integer should returns the default value`)
	}
}

func TestParseIntValue(t *testing.T) {
	if parseInt("2018", 42) != 2018 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestParseListenAddr(t *testing.T) {
	defaultExpected := []string{defaultListenAddr}

	tests := []struct {
		name           string
		listenAddr     string
		port           string
		expected       []string
		lines          []string // Used for direct lines parsing instead of individual env vars
		isLineOriented bool     // Flag to indicate if we use lines
	}{
		{
			name:       "Single LISTEN_ADDR",
			listenAddr: "127.0.0.1:8080",
			expected:   []string{"127.0.0.1:8080"},
		},
		{
			name:       "Multiple LISTEN_ADDR comma-separated",
			listenAddr: "127.0.0.1:8080,:8081,/tmp/miniflux.sock",
			expected:   []string{"127.0.0.1:8080", ":8081", "/tmp/miniflux.sock"},
		},
		{
			name:       "Multiple LISTEN_ADDR with spaces around commas",
			listenAddr: "127.0.0.1:8080 , :8081",
			expected:   []string{"127.0.0.1:8080", ":8081"},
		},
		{
			name:       "Empty LISTEN_ADDR",
			listenAddr: "",
			expected:   defaultExpected,
		},
		{
			name:       "PORT overrides LISTEN_ADDR",
			listenAddr: "127.0.0.1:8000",
			port:       "8082",
			expected:   []string{":8082"},
		},
		{
			name:       "PORT overrides empty LISTEN_ADDR",
			listenAddr: "",
			port:       "8083",
			expected:   []string{":8083"},
		},
		{
			name:       "LISTEN_ADDR with empty segment (comma)",
			listenAddr: "127.0.0.1:8080,,:8081",
			expected:   []string{"127.0.0.1:8080", ":8081"},
		},
		{
			name:           "PORT override with lines parsing",
			isLineOriented: true,
			lines:          []string{"LISTEN_ADDR=127.0.0.1:8000", "PORT=8082"},
			expected:       []string{":8082"},
		},
		{
			name:           "LISTEN_ADDR only with lines parsing (comma)",
			isLineOriented: true,
			lines:          []string{"LISTEN_ADDR=10.0.0.1:9090,10.0.0.2:9091"},
			expected:       []string{"10.0.0.1:9090", "10.0.0.2:9091"},
		},
		{
			name:           "Empty LISTEN_ADDR with lines parsing (default)",
			isLineOriented: true,
			lines:          []string{"LISTEN_ADDR="},
			expected:       defaultExpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			var err error

			if tt.isLineOriented {
				err = parser.parseLines(tt.lines)
			} else {
				// Simulate os.Environ() behaviour for individual var testing
				var envLines []string
				if tt.listenAddr != "" {
					envLines = append(envLines, "LISTEN_ADDR="+tt.listenAddr)
				}
				if tt.port != "" {
					envLines = append(envLines, "PORT="+tt.port)
				}
				// Add a dummy var if both are empty to avoid empty lines slice if not intended
				if tt.listenAddr == "" && tt.port == "" && tt.name == "Empty LISTEN_ADDR" {
					// This case specifically tests empty LISTEN_ADDR resulting in default
					// So, we pass LISTEN_ADDR=
					envLines = append(envLines, "LISTEN_ADDR=")
				}
				err = parser.parseLines(envLines)
			}

			if err != nil {
				t.Fatalf("parseLines() error = %v", err)
			}

			opts := parser.opts
			if !reflect.DeepEqual(opts.ListenAddr(), tt.expected) {
				t.Errorf("ListenAddr() got = %v, want %v", opts.ListenAddr(), tt.expected)
			}
		})
	}
}
