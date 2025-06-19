// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package filter // import "miniflux.app/v2/internal/reader/filter"

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"
)

type filterActionType string

const (
	filterActionBlock filterActionType = "block"
	filterActionAllow filterActionType = "allow"
)

func IsBlockedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	// Check user-defined block rules first
	if user.BlockFilterEntryRules != "" {
		if matchesUserRules(user.BlockFilterEntryRules, entry, feed, filterActionBlock) {
			return true
		}
	}

	// Check feed-level blocklist rules
	if feed.BlocklistRules == "" {
		return false
	}

	return matchesFeedRules(feed.BlocklistRules, entry, feed, filterActionBlock)
}

func IsAllowedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	// Check user-defined keep rules first
	if user.KeepFilterEntryRules != "" {
		return matchesUserRules(user.KeepFilterEntryRules, entry, feed, filterActionAllow)
	}

	// Check feed-level keeplist rules
	if feed.KeeplistRules == "" {
		return true
	}

	return matchesFeedRules(feed.KeeplistRules, entry, feed, filterActionAllow)
}

func matchesUserRules(rules string, entry *model.Entry, feed *model.Feed, filterAction filterActionType) bool {
	for rule := range strings.SplitSeq(rules, "\n") {
		if matchesRule(rule, entry) {
			logFilterAction(entry, feed, rule, filterAction)
			return true
		}
	}
	return false
}

func matchesFeedRules(rules string, entry *model.Entry, feed *model.Feed, filterAction filterActionType) bool {
	compiledRegex, err := regexp.Compile(rules)
	if err != nil {
		slog.Debug("Failed on regexp compilation",
			slog.String("pattern", rules),
			slog.Any("error", err),
		)
		return false
	}

	containsMatchingTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledRegex.MatchString(tag)
	})

	if compiledRegex.MatchString(entry.URL) ||
		compiledRegex.MatchString(entry.Title) ||
		compiledRegex.MatchString(entry.Author) ||
		containsMatchingTag {
		logFilterAction(entry, feed, rules, filterAction)
		return true
	}

	return false
}

func matchesRule(rule string, entry *model.Entry) bool {
	parts := strings.SplitN(rule, "=", 2)
	if len(parts) != 2 {
		return false
	}

	ruleType, ruleValue := parts[0], parts[1]

	switch ruleType {
	case "EntryDate":
		return isDateMatchingPattern(ruleValue, entry.Date)
	case "EntryTitle":
		match, _ := regexp.MatchString(ruleValue, entry.Title)
		return match
	case "EntryURL":
		match, _ := regexp.MatchString(ruleValue, entry.URL)
		return match
	case "EntryCommentsURL":
		match, _ := regexp.MatchString(ruleValue, entry.CommentsURL)
		return match
	case "EntryContent":
		match, _ := regexp.MatchString(ruleValue, entry.Content)
		return match
	case "EntryAuthor":
		match, _ := regexp.MatchString(ruleValue, entry.Author)
		return match
	case "EntryTag":
		return containsRegexPattern(ruleValue, entry.Tags)
	}

	return false
}

func logFilterAction(entry *model.Entry, feed *model.Feed, filterRule string, filterAction filterActionType) {
	slog.Debug("Filtering entry based on rule",
		slog.Int64("feed_id", feed.ID),
		slog.String("feed_url", feed.FeedURL),
		slog.String("entry_url", entry.URL),
		slog.String("filter_rule", filterRule),
		slog.Any("filter_action", filterAction),
	)
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
	}
	return false
}

func containsRegexPattern(pattern string, entries []string) bool {
	for _, entry := range entries {
		if matched, _ := regexp.MatchString(pattern, entry); matched {
			return true
		}
	}
	return false
}
