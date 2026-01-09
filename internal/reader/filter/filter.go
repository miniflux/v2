// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package filter provides functions to filter entries based on user-defined rules.
//
// There are two types of rules:
//
// Block Rules: Ignore articles that match the regex.
// Keep Rules: Retain only articles that match the regex.
//
// Rules are processed in this order:
//
// 1. User block filter rules
// 2. Feed block filter rules
// 3. User keep filter rules
// 4. Feed keep filter rules
//
// Each rule must be on a separate line.
// Duplicate rules are allowed. For example, having multiple EntryTitle rules is possible.
// The provided regex should use the RE2 syntax.
// The order of the rules matters as the processor stops on the first match for both Block and Keep rules.
// Invalid rules are ignored.

package filter // import "miniflux.app/v2/internal/reader/filter"

import (
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"
)

type filterRule struct {
	Type  string
	Value string
}

type filterRules []filterRule

func ParseRules(userRules, feedRules string) filterRules {
	rules := make(filterRules, 0)
	for line := range strings.SplitSeq(strings.TrimSpace(userRules), "\n") {
		if valid, filterRule := parseRule(line); valid {
			rules = append(rules, filterRule)
		}
	}
	for line := range strings.SplitSeq(strings.TrimSpace(feedRules), "\n") {
		if valid, filterRule := parseRule(line); valid {
			rules = append(rules, filterRule)
		}
	}
	return rules
}

func parseRule(userDefinedRule string) (bool, filterRule) {
	userDefinedRule = strings.TrimSpace(strings.ReplaceAll(userDefinedRule, "\r\n", ""))
	parts := strings.SplitN(userDefinedRule, "=", 2)
	if len(parts) != 2 {
		return false, filterRule{}
	}
	return true, filterRule{
		Type:  strings.TrimSpace(parts[0]),
		Value: strings.TrimSpace(parts[1]),
	}
}

func IsBlockedEntry(blockRules filterRules, allowRules filterRules, feed *model.Feed, entry *model.Entry) bool {
	if matchesEntryFilterRules(blockRules, feed, entry) {
		return true
	}

	if matches, valid := matchesEntryRegexRules(feed.BlocklistRules, feed, entry); valid && matches {
		return true
	}

	// If allow rules exist, only entries that match them should be retained
	if len(allowRules) > 0 {
		if !matchesEntryFilterRules(allowRules, feed, entry) {
			return true // Block entry if it doesn't match any allow rules
		}
		return false // Allow entry if it matches allow rules
	}

	// If keeplist rules exist, only entries that match them should be retained
	if feed.KeeplistRules != "" {
		if matches, valid := matchesEntryRegexRules(feed.KeeplistRules, feed, entry); valid && !matches {
			return true // Block entry if it doesn't match keeplist rules
		}
		return false // Allow entry if it matches keeplist rules or rule is invalid (ignored)
	}

	return false
}

// matchesEntryRegexRules checks if the entry matches the regex rules defined in the feed or user settings.
// It returns true if the entry matches the regex pattern, and a boolean indicating if the regex is valid.
func matchesEntryRegexRules(regexPattern string, feed *model.Feed, entry *model.Entry) (bool, bool) {
	if regexPattern == "" {
		return false, true // No pattern means rule is valid but doesn't match
	}

	compiledRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		slog.Warn("Failed on regexp compilation",
			slog.String("regex_pattern", regexPattern),
			slog.Any("error", err),
		)
		return false, false // Invalid regex pattern
	}

	containsMatchingTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledRegex.MatchString(tag)
	})

	if compiledRegex.MatchString(entry.URL) ||
		compiledRegex.MatchString(entry.Title) ||
		compiledRegex.MatchString(entry.Author) ||
		containsMatchingTag {
		slog.Debug("Entry matches regex rule",
			slog.String("entry_url", entry.URL),
			slog.String("entry_title", entry.Title),
			slog.String("entry_author", entry.Author),
			slog.String("feed_url", feed.FeedURL),
			slog.String("regex_pattern", regexPattern),
		)
		return true, true // Pattern matches and is valid
	}

	return false, true // Pattern is valid but doesn't match
}

func matchesEntryFilterRules(rules filterRules, feed *model.Feed, entry *model.Entry) bool {
	for _, rule := range rules {
		if matchesRule(rule, entry) {
			slog.Debug("Entry matches filter rule",
				slog.String("entry_url", entry.URL),
				slog.String("entry_title", entry.Title),
				slog.String("entry_author", entry.Author),
				slog.String("feed_url", feed.FeedURL),
				slog.String("rule_type", rule.Type),
				slog.String("rule_value", rule.Value),
			)
			return true
		}
	}
	return false
}

func matchesRule(rule filterRule, entry *model.Entry) bool {
	switch rule.Type {
	case "EntryDate":
		return isDateMatchingPattern(rule.Value, entry.Date)
	case "EntryTitle":
		match, _ := regexp.MatchString(rule.Value, entry.Title)
		return match
	case "EntryURL":
		match, _ := regexp.MatchString(rule.Value, entry.URL)
		return match
	case "EntryCommentsURL":
		match, _ := regexp.MatchString(rule.Value, entry.CommentsURL)
		return match
	case "EntryContent":
		match, _ := regexp.MatchString(rule.Value, entry.Content)
		return match
	case "EntryAuthor":
		match, _ := regexp.MatchString(rule.Value, entry.Author)
		return match
	case "EntryTag":
		return containsRegexPattern(rule.Value, entry.Tags)
	}

	return false
}

func isDateMatchingPattern(pattern string, entryDate time.Time) bool {
	if pattern == "future" {
		return entryDate.After(time.Now())
	}

	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) != 2 {
		return false
	}

	ruleType, inputDate := parts[0], parts[1]

	switch ruleType {
	case "before":
		targetDate, err := time.Parse("2006-01-02", inputDate)
		if err != nil {
			return false
		}
		return entryDate.Before(targetDate)
	case "after":
		targetDate, err := time.Parse("2006-01-02", inputDate)
		if err != nil {
			return false
		}
		return entryDate.After(targetDate)
	case "between":
		dates := strings.Split(inputDate, ",")
		if len(dates) != 2 {
			return false
		}
		startDate, err := time.Parse("2006-01-02", dates[0])
		if err != nil {
			return false
		}
		endDate, err := time.Parse("2006-01-02", dates[1])
		if err != nil {
			return false
		}
		return entryDate.After(startDate) && entryDate.Before(endDate)
	case "max-age":
		duration, err := parseDuration(inputDate)
		if err != nil {
			return false
		}
		cutoffDate := time.Now().Add(-duration)
		return entryDate.Before(cutoffDate)
	}
	return false
}

func containsRegexPattern(pattern string, items []string) bool {
	for _, item := range items {
		if matched, _ := regexp.MatchString(pattern, item); matched {
			return true
		}
	}
	return false
}

func parseDuration(duration string) (time.Duration, error) {
	// Handle common duration formats like "30d", "7d", "1h", "1m", etc.
	// Go's time.ParseDuration doesn't support days, so we handle them manually
	if strings.HasSuffix(duration, "d") {
		daysStr := strings.TrimSuffix(duration, "d")
		days := 0
		if daysStr != "" {
			var err error
			days, err = strconv.Atoi(daysStr)
			if err != nil {
				return 0, err
			}
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// For other durations (hours, minutes, seconds), use Go's built-in parser
	return time.ParseDuration(duration)
}
