// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestRewriteEntryURL(t *testing.T) {
	scenarios := []struct {
		name        string
		feed        *model.Feed
		entry       *model.Entry
		expectedURL string
		description string
	}{
		{
			name: "NoRewriteRules",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: "",
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when no rewrite rules are specified",
		},
		{
			name: "EmptyRewriteRules",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: "   ",
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when rewrite rules are empty/whitespace",
		},
		{
			name: "ValidRewriteRule",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/article/(.+)"|"https://example.com/full-article/$1")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/full-article/123",
			description: "Should rewrite URL according to the regex pattern",
		},
		{
			name: "ComplexRegexRewrite",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://news.ycombinator.com/rss",
				UrlRewriteRules: `rewrite("^https://news\.ycombinator\.com/item\?id=(.+)"|"https://hn.algolia.com/api/v1/items/$1")`,
			},
			entry: &model.Entry{
				URL: "https://news.ycombinator.com/item?id=12345",
			},
			expectedURL: "https://hn.algolia.com/api/v1/items/12345",
			description: "Should handle complex regex patterns with escaped characters",
		},
		{
			name: "NoMatchingPattern",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://different.com/(.+)"|"https://rewritten.com/$1")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when regex pattern doesn't match",
		},
		{
			name: "InvalidRegexPattern",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/[invalid"|"https://rewritten.com/$1")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when regex pattern is invalid",
		},
		{
			name: "MalformedRewriteRule",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("invalid format")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when rewrite rule format is malformed",
		},
		{
			name: "MultipleGroups",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/([^/]+)/article/(.+)"|"https://example.com/full/$1/story/$2")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/tech/article/ai-news",
			},
			expectedURL: "https://example.com/full/tech/story/ai-news",
			description: "Should handle multiple capture groups in regex",
		},
		{
			name: "URLWithSpecialCharacters",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/(.+)"|"https://proxy.example.com/$1")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/test?param=value&other=123#section",
			},
			expectedURL: "https://proxy.example.com/article/test?param=value&other=123#section",
			description: "Should handle URLs with query parameters and fragments",
		},
		{
			name: "ReplaceWithStaticURL",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/(.+)"|"https://static.example.com/reader")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://static.example.com/reader",
			description: "Should replace with static URL when no capture groups are used in replacement",
		},
		{
			name: "EmptyReplacementString",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/(.+)"|"x")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "x",
			description: "Should replace with specified string",
		},
		{
			name: "EmptyReplacementNotSupported",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `rewrite("^https://example.com/(.+)"|"")`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when replacement is empty string (not supported by regex pattern)",
		},
		{
			name: "InvalidRewriteRuleFormat",
			feed: &model.Feed{
				ID:              1,
				FeedURL:         "https://example.com/feed.xml",
				UrlRewriteRules: `not-a-rewrite-rule`,
			},
			entry: &model.Entry{
				URL: "https://example.com/article/123",
			},
			expectedURL: "https://example.com/article/123",
			description: "Should return original URL when rewrite rule doesn't match expected format",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := RewriteEntryURL(scenario.feed, scenario.entry)
			if result != scenario.expectedURL {
				t.Errorf("Expected URL %q, got %q. Description: %s", scenario.expectedURL, result, scenario.description)
			}
		})
	}
}

func TestRewriteEntryURLWithNilValues(t *testing.T) {
	t.Run("NilFeed", func(t *testing.T) {
		entry := &model.Entry{URL: "https://example.com/article/123"}

		// This should panic or handle gracefully - let's see what happens
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when feed is nil, but function completed normally")
			}
		}()

		RewriteEntryURL(nil, entry)
	})

	t.Run("NilEntry", func(t *testing.T) {
		feed := &model.Feed{
			ID:              1,
			FeedURL:         "https://example.com/feed.xml",
			UrlRewriteRules: `rewrite("^https://example.com/(.+)"|"https://rewritten.com/$1")`,
		}

		// This should panic or handle gracefully - let's see what happens
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when entry is nil, but function completed normally")
			}
		}()

		RewriteEntryURL(feed, nil)
	})
}

func TestCustomReplaceRuleRegex(t *testing.T) {
	scenarios := []struct {
		name     string
		input    string
		expected []string
		matches  bool
	}{
		{
			name:     "ValidRule",
			input:    `rewrite("^https://example.com/(.+)"|"https://rewritten.com/$1")`,
			expected: []string{`rewrite("^https://example.com/(.+)"|"https://rewritten.com/$1")`, `^https://example.com/(.+)`, `https://rewritten.com/$1`},
			matches:  true,
		},
		{
			name:     "ValidRuleWithEscapedCharacters",
			input:    `rewrite("^https://news\\.ycombinator\\.com/item\\?id=(.+)"|"https://hn.algolia.com/api/v1/items/$1")`,
			expected: []string{`rewrite("^https://news\\.ycombinator\\.com/item\\?id=(.+)"|"https://hn.algolia.com/api/v1/items/$1")`, `^https://news\\.ycombinator\\.com/item\\?id=(.+)`, `https://hn.algolia.com/api/v1/items/$1`},
			matches:  true,
		},
		{
			name:     "InvalidFormat",
			input:    `rewrite("invalid")`,
			expected: nil,
			matches:  false,
		},
		{
			name:     "EmptyString",
			input:    ``,
			expected: nil,
			matches:  false,
		},
		{
			name:     "RandomText",
			input:    `some random text`,
			expected: nil,
			matches:  false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			parts := customReplaceRuleRegex.FindStringSubmatch(scenario.input)

			if scenario.matches {
				if len(parts) < 3 {
					t.Errorf("Expected regex to match and return at least 3 parts, got %d parts: %v", len(parts), parts)
					return
				}

				// Check the full match and captured groups
				if parts[0] != scenario.expected[0] {
					t.Errorf("Expected full match %q, got %q", scenario.expected[0], parts[0])
				}
				if parts[1] != scenario.expected[1] {
					t.Errorf("Expected first capture group %q, got %q", scenario.expected[1], parts[1])
				}
				if parts[2] != scenario.expected[2] {
					t.Errorf("Expected second capture group %q, got %q", scenario.expected[2], parts[2])
				}
			} else if len(parts) >= 3 {
				t.Errorf("Expected regex not to match, but got %d parts: %v", len(parts), parts)
			}
		})
	}
}
