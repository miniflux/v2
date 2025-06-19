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

// TODO factorize isBlockedEntry and isAllowedEntry

func IsBlockedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if user.BlockFilterEntryRules != "" {
		rules := strings.SplitSeq(user.BlockFilterEntryRules, "\n")
		for rule := range rules {
			match := false
			parts := strings.SplitN(rule, "=", 2)
			if len(parts) != 2 {
				return false
			}
			part, pattern := parts[0], parts[1]

			switch part {
			case "EntryDate":
				match = isDateMatchingPattern(pattern, entry.Date)
			case "EntryTitle":
				match, _ = regexp.MatchString(pattern, entry.Title)
			case "EntryURL":
				match, _ = regexp.MatchString(pattern, entry.URL)
			case "EntryCommentsURL":
				match, _ = regexp.MatchString(pattern, entry.CommentsURL)
			case "EntryContent":
				match, _ = regexp.MatchString(pattern, entry.Content)
			case "EntryAuthor":
				match, _ = regexp.MatchString(pattern, entry.Author)
			case "EntryTag":
				match = containsRegexPattern(pattern, entry.Tags)
			}

			if match {
				slog.Debug("Blocking entry based on rule",
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.String("rule", rule),
				)
				return true
			}
		}
	}

	if feed.BlocklistRules == "" {
		return false
	}

	compiledBlocklist, err := regexp.Compile(feed.BlocklistRules)
	if err != nil {
		slog.Debug("Failed on regexp compilation",
			slog.String("pattern", feed.BlocklistRules),
			slog.Any("error", err),
		)
		return false
	}

	containsBlockedTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledBlocklist.MatchString(tag)
	})

	if compiledBlocklist.MatchString(entry.URL) || compiledBlocklist.MatchString(entry.Title) || compiledBlocklist.MatchString(entry.Author) || containsBlockedTag {
		slog.Debug("Blocking entry based on rule",
			slog.String("entry_url", entry.URL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
			slog.String("rule", feed.BlocklistRules),
		)
		return true
	}

	return false
}

func IsAllowedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if user.KeepFilterEntryRules != "" {
		rules := strings.SplitSeq(user.KeepFilterEntryRules, "\n")
		for rule := range rules {
			match := false
			parts := strings.SplitN(rule, "=", 2)
			if len(parts) != 2 {
				return false
			}
			part, pattern := parts[0], parts[1]

			switch part {
			case "EntryDate":
				match = isDateMatchingPattern(pattern, entry.Date)
			case "EntryTitle":
				match, _ = regexp.MatchString(pattern, entry.Title)
			case "EntryURL":
				match, _ = regexp.MatchString(pattern, entry.URL)
			case "EntryCommentsURL":
				match, _ = regexp.MatchString(pattern, entry.CommentsURL)
			case "EntryContent":
				match, _ = regexp.MatchString(pattern, entry.Content)
			case "EntryAuthor":
				match, _ = regexp.MatchString(pattern, entry.Author)
			case "EntryTag":
				match = containsRegexPattern(pattern, entry.Tags)
			}

			if match {
				slog.Debug("Allowing entry based on rule",
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.String("rule", rule),
				)
				return true
			}
		}
		return false
	}

	if feed.KeeplistRules == "" {
		return true
	}

	compiledKeeplist, err := regexp.Compile(feed.KeeplistRules)
	if err != nil {
		slog.Debug("Failed on regexp compilation",
			slog.String("pattern", feed.KeeplistRules),
			slog.Any("error", err),
		)
		return false
	}
	containsAllowedTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledKeeplist.MatchString(tag)
	})

	if compiledKeeplist.MatchString(entry.URL) || compiledKeeplist.MatchString(entry.Title) || compiledKeeplist.MatchString(entry.Author) || containsAllowedTag {
		slog.Debug("Allow entry based on rule",
			slog.String("entry_url", entry.URL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
			slog.String("rule", feed.KeeplistRules),
		)
		return true
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

	operator, dateStr := parts[0], parts[1]

	switch operator {
	case "before":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.Before(targetDate)
	case "after":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.After(targetDate)
	case "between":
		dates := strings.Split(dateStr, ",")
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
