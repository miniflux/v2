// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package filter // import "miniflux.app/v2/internal/reader/filter"

import (
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

// Test helper functions
func createTestEntry() *model.Entry {
	return &model.Entry{
		ID:          1,
		Title:       "Test Entry Title",
		URL:         "https://example.com/test-entry",
		CommentsURL: "https://example.com/test-entry/comments",
		Content:     "This is the test entry content",
		Author:      "Test Author",
		Date:        time.Now(),
		Tags:        []string{"golang", "testing", "miniflux"},
	}
}

func createTestFeed() *model.Feed {
	return &model.Feed{
		ID:                    1,
		FeedURL:               "https://example.com/feed.xml",
		BlocklistRules:        "",
		KeeplistRules:         "",
		BlockFilterEntryRules: "",
		KeepFilterEntryRules:  "",
	}
}

// Tests for ParseRules function
func TestParseRules(t *testing.T) {
	tests := []struct {
		name      string
		userRules string
		feedRules string
		expected  int
	}{
		{
			name:      "empty rules",
			userRules: "",
			feedRules: "",
			expected:  0,
		},
		{
			name:      "valid user rules only",
			userRules: "EntryTitle=test\nEntryAuthor=author",
			feedRules: "",
			expected:  2,
		},
		{
			name:      "valid feed rules only",
			userRules: "",
			feedRules: "EntryURL=example\nEntryContent=content",
			expected:  2,
		},
		{
			name:      "both user and feed rules",
			userRules: "EntryTitle=test\nEntryAuthor=author",
			feedRules: "EntryURL=example\nEntryContent=content",
			expected:  4,
		},
		{
			name:      "mixed valid and invalid rules",
			userRules: "EntryTitle=test\ninvalid_rule\nEntryAuthor=author",
			feedRules: "EntryURL=example\nanotherInvalid\nEntryContent=content",
			expected:  4,
		},
		{
			name:      "rules with carriage returns",
			userRules: "EntryTitle=test\r\nEntryAuthor=author\r\n",
			feedRules: "",
			expected:  2,
		},
		{
			name:      "rules with extra whitespace",
			userRules: "  EntryTitle  =  test  \n  EntryAuthor  =  author  ",
			feedRules: "",
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := ParseRules(tt.userRules, tt.feedRules)
			if len(rules) != tt.expected {
				t.Errorf("ParseRules() returned %d rules, expected %d", len(rules), tt.expected)
			}
		})
	}
}

// Tests for parseRule function
func TestParseRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		valid    bool
		expected filterRule
	}{
		{
			name:     "valid rule",
			rule:     "EntryTitle=test",
			valid:    true,
			expected: filterRule{Type: "EntryTitle", Value: "test"},
		},
		{
			name:     "rule with extra whitespace",
			rule:     "  EntryTitle  =  test  ",
			valid:    true,
			expected: filterRule{Type: "EntryTitle", Value: "test"},
		},
		{
			name:     "rule with carriage return",
			rule:     "EntryTitle=test\r\n",
			valid:    true,
			expected: filterRule{Type: "EntryTitle", Value: "test"},
		},
		{
			name:     "rule with single carriage return",
			rule:     "EntryTitle=test\r",
			valid:    true,
			expected: filterRule{Type: "EntryTitle", Value: "test"},
		},
		{
			name:  "invalid rule - no equals",
			rule:  "EntryTitle",
			valid: false,
		},
		{
			name:  "invalid rule - empty",
			rule:  "",
			valid: false,
		},
		{
			name:     "invalid rule - multiple equals",
			rule:     "EntryTitle=test=value",
			valid:    true,
			expected: filterRule{Type: "EntryTitle", Value: "test=value"},
		},
		{
			name:     "rule with equals in value",
			rule:     "EntryContent=x=y",
			valid:    true,
			expected: filterRule{Type: "EntryContent", Value: "x=y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, rule := parseRule(tt.rule)
			if valid != tt.valid {
				t.Errorf("parseRule() validity = %v, expected %v", valid, tt.valid)
			}
			if valid && (rule.Type != tt.expected.Type || rule.Value != tt.expected.Value) {
				t.Errorf("parseRule() = %+v, expected %+v", rule, tt.expected)
			}
		})
	}
}

// Tests for IsBlockedEntry function
func TestIsBlockedEntry(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name       string
		blockRules filterRules
		allowRules filterRules
		setup      func()
		expected   bool
	}{
		{
			name:       "no rules - not blocked",
			blockRules: filterRules{},
			allowRules: filterRules{},
			setup:      func() {},
			expected:   false,
		},
		{
			name:       "matching block rule",
			blockRules: filterRules{{Type: "EntryTitle", Value: "Test"}},
			allowRules: filterRules{},
			setup:      func() {},
			expected:   true,
		},
		{
			name:       "block rule takes precedence over allow rule",
			blockRules: filterRules{{Type: "EntryTitle", Value: "Test"}},
			allowRules: filterRules{{Type: "EntryTitle", Value: "Test"}},
			setup:      func() {},
			expected:   true, // Block rules are checked first
		},
		{
			name:       "non-matching block rule",
			blockRules: filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			allowRules: filterRules{},
			setup:      func() {},
			expected:   false,
		},
		{
			name:       "allow rule matches - entry should be allowed",
			blockRules: filterRules{},
			allowRules: filterRules{{Type: "EntryTitle", Value: "Test"}},
			setup:      func() {},
			expected:   false,
		},
		{
			name:       "allow rule exists but doesn't match - entry should be blocked",
			blockRules: filterRules{},
			allowRules: filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			setup:      func() {},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := IsBlockedEntry(tt.blockRules, tt.allowRules, feed, entry)
			if result != tt.expected {
				t.Errorf("IsBlockedEntry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAllowRulesExclusiveBehavior(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name        string
		allowRules  filterRules
		expected    bool
		description string
	}{
		{
			name:        "no allow rules - entry should pass",
			allowRules:  filterRules{},
			expected:    false,
			description: "When no allow rules exist, entry should not be blocked",
		},
		{
			name:        "allow rule matches - entry should pass",
			allowRules:  filterRules{{Type: "EntryTitle", Value: "Test"}},
			expected:    false,
			description: "When allow rules exist and match, entry should not be blocked",
		},
		{
			name:        "allow rule doesn't match - entry should be blocked",
			allowRules:  filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			expected:    true,
			description: "When allow rules exist but don't match, entry should be blocked",
		},
		{
			name: "multiple allow rules - one matches",
			allowRules: filterRules{
				{Type: "EntryTitle", Value: "NonMatching"},
				{Type: "EntryAuthor", Value: "Test"},
			},
			expected:    false,
			description: "When any allow rule matches, entry should not be blocked",
		},
		{
			name: "multiple allow rules - none match",
			allowRules: filterRules{
				{Type: "EntryTitle", Value: "NonMatching1"},
				{Type: "EntryAuthor", Value: "NonMatching2"},
			},
			expected:    true,
			description: "When no allow rules match, entry should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBlockedEntry(filterRules{}, tt.allowRules, feed, entry)
			if result != tt.expected {
				t.Errorf("IsBlockedEntry() = %v, expected %v (%s)", result, tt.expected, tt.description)
			}
		})
	}
}

func TestAllowRulesWithBlockRulesPrecedence(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name        string
		blockRules  filterRules
		allowRules  filterRules
		expected    bool
		description string
	}{
		{
			name:        "block rule takes precedence over matching allow rule",
			blockRules:  filterRules{{Type: "EntryTitle", Value: "Test"}},
			allowRules:  filterRules{{Type: "EntryTitle", Value: "Test"}},
			expected:    true,
			description: "Block rules should always take precedence, even when allow rules would match",
		},
		{
			name:        "block rule takes precedence, allow rule would fail anyway",
			blockRules:  filterRules{{Type: "EntryTitle", Value: "Test"}},
			allowRules:  filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			expected:    true,
			description: "Block rules should take precedence regardless of allow rule matching",
		},
		{
			name:        "no block rule, allow rule matches",
			blockRules:  filterRules{},
			allowRules:  filterRules{{Type: "EntryTitle", Value: "Test"}},
			expected:    false,
			description: "When no block rules match and allow rule matches, entry should pass",
		},
		{
			name:        "non-matching block rule, allow rule doesn't match",
			blockRules:  filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			allowRules:  filterRules{{Type: "EntryTitle", Value: "NonMatching"}},
			expected:    true,
			description: "When block rules don't match but allow rules also don't match, entry should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBlockedEntry(tt.blockRules, tt.allowRules, feed, entry)
			if result != tt.expected {
				t.Errorf("IsBlockedEntry() = %v, expected %v (%s)", result, tt.expected, tt.description)
			}
		})
	}
}

func TestKeeplistRulesBehavior(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name         string
		keeplistRule string
		expected     bool
		description  string
	}{
		{
			name:         "no keeplist rules - entry should pass",
			keeplistRule: "",
			expected:     false,
			description:  "When no keeplist rules exist, entry should not be blocked",
		},
		{
			name:         "keeplist rule matches title - entry should pass",
			keeplistRule: "Test.*Title",
			expected:     false,
			description:  "When keeplist rule matches entry title, entry should not be blocked",
		},
		{
			name:         "keeplist rule matches URL - entry should pass",
			keeplistRule: "example\\.com",
			expected:     false,
			description:  "When keeplist rule matches entry URL, entry should not be blocked",
		},
		{
			name:         "keeplist rule matches author - entry should pass",
			keeplistRule: "Test.*Author",
			expected:     false,
			description:  "When keeplist rule matches entry author, entry should not be blocked",
		},
		{
			name:         "keeplist rule matches tag - entry should pass",
			keeplistRule: "golang",
			expected:     false,
			description:  "When keeplist rule matches entry tag, entry should not be blocked",
		},
		{
			name:         "keeplist rule doesn't match - entry should be blocked",
			keeplistRule: "NonMatchingPattern",
			expected:     true,
			description:  "When keeplist rule doesn't match any entry field, entry should be blocked",
		},
		{
			name:         "invalid keeplist regex - entry should pass",
			keeplistRule: "[invalid",
			expected:     false,
			description:  "When keeplist rule is invalid regex, entry should not be blocked (rule is ignored)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed.KeeplistRules = tt.keeplistRule
			feed.BlocklistRules = "" // Ensure no blocklist interference
			result := IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
			if result != tt.expected {
				t.Errorf("IsBlockedEntry() with keeplist '%s' = %v, expected %v (%s)",
					tt.keeplistRule, result, tt.expected, tt.description)
			}
		})
	}
}

// Tests for matchesEntryRegexRules function
func TestMatchesEntryRegexRules(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name          string
		regexPattern  string
		expectedMatch bool
		expectedValid bool
		description   string
	}{
		{
			name:          "empty pattern",
			regexPattern:  "",
			expectedMatch: false,
			expectedValid: true,
			description:   "Empty pattern should be valid but not match",
		},
		{
			name:          "invalid regex",
			regexPattern:  "[",
			expectedMatch: false,
			expectedValid: false,
			description:   "Invalid regex should return false for both match and validity",
		},
		{
			name:          "matches title",
			regexPattern:  "Test.*Title",
			expectedMatch: true,
			expectedValid: true,
			description:   "Valid regex matching title should return true for both",
		},
		{
			name:          "matches URL",
			regexPattern:  "example\\.com",
			expectedMatch: true,
			expectedValid: true,
			description:   "Valid regex matching URL should return true for both",
		},
		{
			name:          "matches author",
			regexPattern:  "Test.*Author",
			expectedMatch: true,
			expectedValid: true,
			description:   "Valid regex matching author should return true for both",
		},
		{
			name:          "matches tag",
			regexPattern:  "golang",
			expectedMatch: true,
			expectedValid: true,
			description:   "Valid regex matching tag should return true for both",
		},
		{
			name:          "no match but valid regex",
			regexPattern:  "nomatch",
			expectedMatch: false,
			expectedValid: true,
			description:   "Valid regex with no match should return false for match, true for validity",
		},
		{
			name:          "invalid regex - unclosed parenthesis",
			regexPattern:  "(unclosed",
			expectedMatch: false,
			expectedValid: false,
			description:   "Invalid regex with unclosed parenthesis should return false for both",
		},
		{
			name:          "invalid regex - invalid quantifier",
			regexPattern:  "*invalid",
			expectedMatch: false,
			expectedValid: false,
			description:   "Invalid regex with wrong quantifier should return false for both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, valid := matchesEntryRegexRules(tt.regexPattern, feed, entry)
			if match != tt.expectedMatch {
				t.Errorf("matchesEntryRegexRules() match = %v, expected %v (%s)", match, tt.expectedMatch, tt.description)
			}
			if valid != tt.expectedValid {
				t.Errorf("matchesEntryRegexRules() valid = %v, expected %v (%s)", valid, tt.expectedValid, tt.description)
			}
		})
	}
}

// Tests for matchesEntryFilterRules function
func TestMatchesEntryFilterRules(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	tests := []struct {
		name     string
		rules    filterRules
		expected bool
	}{
		{
			name:     "empty rules",
			rules:    filterRules{},
			expected: false,
		},
		{
			name: "matching rule",
			rules: filterRules{
				{Type: "EntryTitle", Value: "Test"},
			},
			expected: true,
		},
		{
			name: "non-matching rule",
			rules: filterRules{
				{Type: "EntryTitle", Value: "NonMatching"},
			},
			expected: false,
		},
		{
			name: "multiple rules - one matches",
			rules: filterRules{
				{Type: "EntryTitle", Value: "NonMatching"},
				{Type: "EntryAuthor", Value: "Test"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesEntryFilterRules(tt.rules, feed, entry)
			if result != tt.expected {
				t.Errorf("matchesEntryFilterRules() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Tests for matchesRule function
func TestMatchesRule(t *testing.T) {
	entry := createTestEntry()
	futureEntry := createTestEntry()
	futureEntry.Date = time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		rule     filterRule
		entry    *model.Entry
		expected bool
	}{
		{
			name:     "EntryTitle match",
			rule:     filterRule{Type: "EntryTitle", Value: "Test"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryTitle no match",
			rule:     filterRule{Type: "EntryTitle", Value: "NoMatch"},
			entry:    entry,
			expected: false,
		},
		{
			name:     "EntryURL match",
			rule:     filterRule{Type: "EntryURL", Value: "example\\.com"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryURL no match",
			rule:     filterRule{Type: "EntryURL", Value: "nomatch\\.com"},
			entry:    entry,
			expected: false,
		},
		{
			name:     "EntryCommentsURL match",
			rule:     filterRule{Type: "EntryCommentsURL", Value: "comments"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryContent match",
			rule:     filterRule{Type: "EntryContent", Value: "test.*content"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryAuthor match",
			rule:     filterRule{Type: "EntryAuthor", Value: "Test.*Author"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryTag match",
			rule:     filterRule{Type: "EntryTag", Value: "golang"},
			entry:    entry,
			expected: true,
		},
		{
			name:     "EntryTag no match",
			rule:     filterRule{Type: "EntryTag", Value: "python"},
			entry:    entry,
			expected: false,
		},
		{
			name:     "EntryDate future",
			rule:     filterRule{Type: "EntryDate", Value: "future"},
			entry:    futureEntry,
			expected: true,
		},
		{
			name:     "EntryDate not future",
			rule:     filterRule{Type: "EntryDate", Value: "future"},
			entry:    entry,
			expected: false,
		},
		{
			name:     "unknown rule type",
			rule:     filterRule{Type: "UnknownType", Value: "test"},
			entry:    entry,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesRule(tt.rule, tt.entry)
			if result != tt.expected {
				t.Errorf("matchesRule() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Tests for isDateMatchingPattern function
func TestIsDateMatchingPattern(t *testing.T) {
	now := time.Now()
	testDate := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		pattern   string
		entryDate time.Time
		expected  bool
	}{
		{
			name:      "future - positive case",
			pattern:   "future",
			entryDate: now.Add(time.Hour),
			expected:  true,
		},
		{
			name:      "future - negative case",
			pattern:   "future",
			entryDate: now.Add(-time.Hour),
			expected:  false,
		},
		{
			name:      "before - positive case",
			pattern:   "before:2023-07-01",
			entryDate: testDate,
			expected:  true,
		},
		{
			name:      "before - negative case",
			pattern:   "before:2023-06-01",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "before - invalid date",
			pattern:   "before:invalid-date",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "after - positive case",
			pattern:   "after:2023-06-01",
			entryDate: testDate,
			expected:  true,
		},
		{
			name:      "after - negative case",
			pattern:   "after:2023-07-01",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "after - invalid date",
			pattern:   "after:invalid-date",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "between - positive case",
			pattern:   "between:2023-06-01,2023-07-01",
			entryDate: testDate,
			expected:  true,
		},
		{
			name:      "between - negative case",
			pattern:   "between:2023-07-01,2023-08-01",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "between - invalid format",
			pattern:   "between:2023-06-01",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "between - invalid start date",
			pattern:   "between:invalid,2023-07-01",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "between - invalid end date",
			pattern:   "between:2023-06-01,invalid",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "max-age - positive case",
			pattern:   "max-age:1d",
			entryDate: now.Add(-2 * 24 * time.Hour),
			expected:  true,
		},
		{
			name:      "max-age - negative case",
			pattern:   "max-age:3d",
			entryDate: now.Add(-2 * 24 * time.Hour),
			expected:  false,
		},
		{
			name:      "max-age - invalid duration",
			pattern:   "max-age:invalid",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "invalid pattern format",
			pattern:   "invalid-pattern",
			entryDate: testDate,
			expected:  false,
		},
		{
			name:      "unknown rule type",
			pattern:   "unknown:value",
			entryDate: testDate,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateMatchingPattern(tt.pattern, tt.entryDate)
			if result != tt.expected {
				t.Errorf("isDateMatchingPattern() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Tests for containsRegexPattern function
func TestContainsRegexPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		items    []string
		expected bool
	}{
		{
			name:     "match found",
			pattern:  "go.*",
			items:    []string{"golang", "python", "javascript"},
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "rust",
			items:    []string{"golang", "python", "javascript"},
			expected: false,
		},
		{
			name:     "empty items",
			pattern:  "test",
			items:    []string{},
			expected: false,
		},
		{
			name:     "invalid regex",
			pattern:  "[",
			items:    []string{"test"},
			expected: false,
		},
		{
			name:     "case sensitive match",
			pattern:  "Go",
			items:    []string{"golang", "python"},
			expected: false,
		},
		{
			name:     "exact match",
			pattern:  "^golang$",
			items:    []string{"golang", "go"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsRegexPattern(tt.pattern, tt.items)
			if result != tt.expected {
				t.Errorf("containsRegexPattern() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Tests for parseDuration function
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "days - single digit",
			duration:    "1d",
			expected:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "days - multiple digits",
			duration:    "30d",
			expected:    30 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "days - zero",
			duration:    "0d",
			expected:    0,
			expectError: false,
		},
		{
			name:        "days - empty number",
			duration:    "d",
			expected:    0,
			expectError: false,
		},
		{
			name:        "days - invalid number",
			duration:    "invalid_d",
			expected:    0,
			expectError: true,
		},
		{
			name:        "hours",
			duration:    "24h",
			expected:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "minutes",
			duration:    "60m",
			expected:    60 * time.Minute,
			expectError: false,
		},
		{
			name:        "seconds",
			duration:    "30s",
			expected:    30 * time.Second,
			expectError: false,
		},
		{
			name:        "milliseconds",
			duration:    "500ms",
			expected:    500 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "microseconds",
			duration:    "1000us",
			expected:    1000 * time.Microsecond,
			expectError: false,
		},
		{
			name:        "nanoseconds",
			duration:    "1000ns",
			expected:    1000 * time.Nanosecond,
			expectError: false,
		},
		{
			name:        "invalid duration",
			duration:    "invalid",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			duration:    "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.duration)
			if tt.expectError && err == nil {
				t.Errorf("parseDuration() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("parseDuration() unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("parseDuration() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Additional edge case tests
func TestParseRulesEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		userRules string
		feedRules string
		expected  int
	}{
		{
			name:      "rules with only newlines",
			userRules: "\n\n\n",
			feedRules: "\n\n",
			expected:  0,
		},
		{
			name:      "rules with only whitespace",
			userRules: "   \n   \t   \n",
			feedRules: "",
			expected:  0,
		},
		{
			name:      "rules with equals but empty value",
			userRules: "EntryTitle=",
			feedRules: "",
			expected:  1,
		},
		{
			name:      "rules with equals but empty key",
			userRules: "=value",
			feedRules: "",
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := ParseRules(tt.userRules, tt.feedRules)
			if len(rules) != tt.expected {
				t.Errorf("ParseRules() returned %d rules, expected %d", len(rules), tt.expected)
			}
		})
	}
}

func TestIsBlockedEntryWithRegexRules(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	// Test with blocklist regex rules
	feed.BlocklistRules = "Test.*Title"
	result := IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
	if !result {
		t.Errorf("IsBlockedEntry() should block entry matching blocklist regex")
	}

	// Test with both blocklist and keeplist regex rules - blocklist takes precedence
	feed.KeeplistRules = "Test.*Title"
	result = IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
	if !result {
		t.Errorf("IsBlockedEntry() should block entry when both blocklist and keeplist match (blocklist takes precedence)")
	}

	// Reset blocklist and test with keeplist only
	feed.BlocklistRules = ""
	feed.KeeplistRules = "Test.*Title"
	result = IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
	if result {
		t.Errorf("IsBlockedEntry() should not block entry matching keeplist only")
	}

	// Test with keeplist that doesn't match - should block
	feed.KeeplistRules = "NonMatchingPattern"
	result = IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
	if !result {
		t.Errorf("IsBlockedEntry() should block entry when keeplist doesn't match")
	}
}

func TestMatchesRuleWithInvalidRegex(t *testing.T) {
	entry := createTestEntry()

	// Test invalid regex patterns
	rule := filterRule{Type: "EntryTitle", Value: "["}
	result := matchesRule(rule, entry)
	if result {
		t.Errorf("matchesRule() should return false for invalid regex")
	}
}

func TestIsDateMatchingPatternEdgeCases(t *testing.T) {
	testDate := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	// Test edge case: between with boundary dates
	result := isDateMatchingPattern("between:2023-06-15,2023-06-15", testDate)
	if result {
		t.Errorf("isDateMatchingPattern() should return false for date exactly on boundaries")
	}

	// Test edge case: max-age with hours
	now := time.Now()
	oldEntry := now.Add(-25 * time.Hour)
	result = isDateMatchingPattern("max-age:24h", oldEntry)
	if !result {
		t.Errorf("isDateMatchingPattern() should match old entry with max-age in hours")
	}
}

// Additional comprehensive edge case tests
func TestComplexFilterScenarios(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	// Test complex scenario: block filter rules + blocklist regex + allow filter rules + keeplist regex
	blockRules := filterRules{{Type: "EntryAuthor", Value: "Test.*Author"}}
	allowRules := filterRules{{Type: "EntryTitle", Value: "Test.*Title"}}
	feed.BlocklistRules = "golang"
	feed.KeeplistRules = "testing"

	// Block filter rules should take precedence
	result := IsBlockedEntry(blockRules, allowRules, feed, entry)
	if !result {
		t.Errorf("Complex scenario: block filter rules should take precedence")
	}

	// Remove block filter rules, now blocklist regex should block
	result = IsBlockedEntry(filterRules{}, allowRules, feed, entry)
	if !result {
		t.Errorf("Complex scenario: blocklist regex should block when no filter block rules")
	}

	// Remove blocklist regex, allow filter rules should allow (since they match)
	feed.BlocklistRules = ""
	result = IsBlockedEntry(filterRules{}, allowRules, feed, entry)
	if result {
		t.Errorf("Complex scenario: allow filter rules should not block when they match")
	}

	// Change allow filter rules to non-matching, should block
	allowRules = filterRules{{Type: "EntryTitle", Value: "NonMatching"}}
	result = IsBlockedEntry(filterRules{}, allowRules, feed, entry)
	if !result {
		t.Errorf("Complex scenario: non-matching allow filter rules should block")
	}

	// Remove allow filter rules, keeplist regex should allow
	result = IsBlockedEntry(filterRules{}, filterRules{}, feed, entry)
	if result {
		t.Errorf("Complex scenario: keeplist regex should not block when it matches")
	}
}

func TestFilterRulesWithSpecialCharacters(t *testing.T) {
	entry := &model.Entry{
		Title:   "Test [Special] (Characters) & Symbols!",
		URL:     "https://example.com/test?param=value&other=123",
		Content: "Content with <html> tags and $pecial characters",
		Author:  "Author@domain.com",
		Tags:    []string{"c++", "c#", ".net"},
	}

	tests := []struct {
		name     string
		rule     filterRule
		expected bool
	}{
		{
			name:     "brackets in title",
			rule:     filterRule{Type: "EntryTitle", Value: "\\[Special\\]"},
			expected: true,
		},
		{
			name:     "parentheses in title",
			rule:     filterRule{Type: "EntryTitle", Value: "\\(Characters\\)"},
			expected: true,
		},
		{
			name:     "URL with query parameters",
			rule:     filterRule{Type: "EntryURL", Value: "param=value"},
			expected: true,
		},
		{
			name:     "HTML tags in content",
			rule:     filterRule{Type: "EntryContent", Value: "<html>"},
			expected: true,
		},
		{
			name:     "email pattern in author",
			rule:     filterRule{Type: "EntryAuthor", Value: "@domain\\.com"},
			expected: true,
		},
		{
			name:     "programming language tags",
			rule:     filterRule{Type: "EntryTag", Value: "c\\+\\+"},
			expected: true,
		},
		{
			name:     "tags with special chars",
			rule:     filterRule{Type: "EntryTag", Value: "c#"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesRule(tt.rule, entry)
			if result != tt.expected {
				t.Errorf("matchesRule() with special characters = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEntryWithEmptyFields(t *testing.T) {
	entry := &model.Entry{
		Title:       "",
		URL:         "",
		CommentsURL: "",
		Content:     "",
		Author:      "",
		Tags:        []string{},
		Date:        time.Time{}, // Zero time
	}

	tests := []struct {
		name     string
		rule     filterRule
		expected bool
	}{
		{
			name:     "empty title",
			rule:     filterRule{Type: "EntryTitle", Value: ".*"},
			expected: true, // Empty string matches .*
		},
		{
			name:     "empty title specific match",
			rule:     filterRule{Type: "EntryTitle", Value: "^$"},
			expected: true, // Empty string matches ^$
		},
		{
			name:     "empty URL",
			rule:     filterRule{Type: "EntryURL", Value: "^$"},
			expected: true,
		},
		{
			name:     "empty tags",
			rule:     filterRule{Type: "EntryTag", Value: "anything"},
			expected: false, // No tags to match
		},
		{
			name:     "zero time as future",
			rule:     filterRule{Type: "EntryDate", Value: "future"},
			expected: false, // Zero time is not in future
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesRule(tt.rule, entry)
			if result != tt.expected {
				t.Errorf("matchesRule() with empty fields = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBoundaryConditionsForDates(t *testing.T) {
	// Test dates at exact boundaries
	exactDate := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		pattern   string
		entryDate time.Time
		expected  bool
	}{
		{
			name:      "exact boundary - before same date",
			pattern:   "before:2023-06-15",
			entryDate: exactDate,
			expected:  false,
		},
		{
			name:      "exact boundary - after same date",
			pattern:   "after:2023-06-15",
			entryDate: exactDate,
			expected:  false,
		},
		{
			name:      "one second before boundary",
			pattern:   "before:2023-06-15",
			entryDate: exactDate.Add(-time.Second),
			expected:  true,
		},
		{
			name:      "one second after boundary",
			pattern:   "after:2023-06-15",
			entryDate: exactDate.Add(time.Second),
			expected:  true,
		},
		{
			name:      "between same dates",
			pattern:   "between:2023-06-15,2023-06-15",
			entryDate: exactDate,
			expected:  false, // Entry is not between identical dates
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateMatchingPattern(tt.pattern, tt.entryDate)
			if result != tt.expected {
				t.Errorf("isDateMatchingPattern() boundary test = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRegexErrorHandling(t *testing.T) {
	entry := createTestEntry()
	feed := createTestFeed()

	// Test invalid regex in various contexts
	tests := []struct {
		name         string
		regexPattern string
		expected     bool
	}{
		{
			name:         "invalid regex - unclosed bracket",
			regexPattern: "[abc",
			expected:     false,
		},
		{
			name:         "invalid regex - unclosed parenthesis",
			regexPattern: "(abc",
			expected:     false,
		},
		{
			name:         "invalid regex - invalid quantifier",
			regexPattern: "*abc",
			expected:     false,
		},
		{
			name:         "valid complex regex",
			regexPattern: "^Test.*Entry.*Title$",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := matchesEntryRegexRules(tt.regexPattern, feed, entry)
			if result != tt.expected {
				t.Errorf("matchesEntryRegexRules() with invalid regex = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParseDurationWithVariousFormats(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		expected    time.Duration
		expectError bool
	}{
		// Additional duration format tests
		{
			name:        "complex duration - hours and minutes",
			duration:    "1h30m",
			expected:    time.Hour + 30*time.Minute,
			expectError: false,
		},
		{
			name:        "complex duration - minutes and seconds",
			duration:    "30m45s",
			expected:    30*time.Minute + 45*time.Second,
			expectError: false,
		},
		{
			name:        "fractional hours",
			duration:    "1.5h",
			expected:    time.Hour + 30*time.Minute,
			expectError: false,
		},
		{
			name:        "negative duration",
			duration:    "-1h",
			expected:    -time.Hour,
			expectError: false,
		},
		{
			name:        "zero duration",
			duration:    "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "large number of days",
			duration:    "999d",
			expected:    999 * 24 * time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.duration)
			if tt.expectError && err == nil {
				t.Errorf("parseDuration() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("parseDuration() unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("parseDuration() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkParseRules(b *testing.B) {
	userRules := `EntryTitle=test1
EntryAuthor=author1
EntryURL=example1
EntryContent=content1
EntryTag=tag1`
	feedRules := `EntryTitle=test2
EntryAuthor=author2
EntryURL=example2
EntryContent=content2
EntryTag=tag2`

	b.ResetTimer()
	for b.Loop() {
		ParseRules(userRules, feedRules)
	}
}

func BenchmarkIsBlockedEntry(b *testing.B) {
	entry := createTestEntry()
	feed := createTestFeed()
	blockRules := filterRules{
		{Type: "EntryTitle", Value: "test"},
		{Type: "EntryAuthor", Value: "author"},
		{Type: "EntryURL", Value: "example"},
	}
	allowRules := filterRules{
		{Type: "EntryContent", Value: "content"},
		{Type: "EntryTag", Value: "tag"},
	}

	for b.Loop() {
		IsBlockedEntry(blockRules, allowRules, feed, entry)
	}
}

func BenchmarkMatchesEntryRegexRules(b *testing.B) {
	entry := createTestEntry()
	feed := createTestFeed()
	regexPattern := "Test.*Title|example\\.com|Test.*Author|golang"

	for b.Loop() {
		matchesEntryRegexRules(regexPattern, feed, entry)
	}
}

func BenchmarkIsDateMatchingPattern(b *testing.B) {
	entryDate := time.Now().Add(-2 * 24 * time.Hour)
	pattern := "max-age:1d"

	for b.Loop() {
		isDateMatchingPattern(pattern, entryDate)
	}
}

func BenchmarkParseDuration(b *testing.B) {
	for b.Loop() {
		parseDuration("30d")
	}
}
