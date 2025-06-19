// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestIsRecentEntry(t *testing.T) {
	parser := config.NewParser()
	var err error
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}
	var scenarios = []struct {
		entry    *model.Entry
		expected bool
	}{
		{&model.Entry{Title: "Example1", Date: time.Date(2005, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example2", Date: time.Date(2010, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example3", Date: time.Date(2020, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example4", Date: time.Date(2024, 3, 15, 05, 05, 05, 05, time.UTC)}, true},
	}
	for _, tc := range scenarios {
		result := isRecentEntry(tc.entry)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

func TestMinifyEntryContent(t *testing.T) {
	input := `<p>    Some text with a <a href="http://example.org/"> link   </a>    </p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a></p>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}
