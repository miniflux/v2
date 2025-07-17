// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package filter // import "miniflux.app/v2/internal/reader/filter"

import (
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

type filterActionType string

const (
	filterActionBlock filterActionType = "block"
	filterActionAllow filterActionType = "allow"
)

func isBlockedGlobally(entry *model.Entry) bool {
	if config.Opts == nil {
		return false
	}

	if config.Opts.FilterEntryMaxAgeDays() > 0 {
		maxAge := time.Duration(config.Opts.FilterEntryMaxAgeDays()) * 24 * time.Hour
		if entry.Date.Before(time.Now().Add(-maxAge)) {
			slog.Debug("Entry is blocked globally due to max age",
				slog.String("entry_url", entry.URL),
				slog.Time("entry_date", entry.Date),
				slog.Duration("max_age", maxAge),
			)
			return true
		}
	}

	return false
}

func IsBlockedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if isBlockedGlobally(entry) {
		return true
	}

	combinedRules := combineFilterRules(user.BlockFilterEntryRules, feed.BlockFilterEntryRules)
	if combinedRules != "" {
		if matchesEntryFilterRules(combinedRules, entry, feed, filterActionBlock) {
			return true
		}
	}

	if feed.BlocklistRules == "" {
		return false
	}

	return matchesEntryRegexRules(feed.BlocklistRules, entry, feed, filterActionBlock)
}

func IsAllowedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	combinedRules := combineFilterRules(user.KeepFilterEntryRules, feed.KeepFilterEntryRules)
	if combinedRules != "" {
		return matchesEntryFilterRules(combinedRules, entry, feed, filterActionAllow)
	}

	if feed.KeeplistRules == "" {
		return true
	}

	return matchesEntryRegexRules(feed.KeeplistRules, entry, feed, filterActionAllow)
}

func combineFilterRules(userRules, feedRules string) string {
	var combinedRules strings.Builder

	userRules = strings.TrimSpace(userRules)
	feedRules = strings.TrimSpace(feedRules)

	if userRules != "" {
		combinedRules.WriteString(userRules)
	}
	if feedRules != "" {
		if combinedRules.Len() > 0 {
			combinedRules.WriteString("\n")
		}
		combinedRules.WriteString(feedRules)
	}
	return combinedRules.String()
}

func matchesEntryFilterRules(rules string, entry *model.Entry, feed *model.Feed, filterAction filterActionType) bool {
	for rule := range strings.SplitSeq(rules, "\n") {
		if matchesRule(rule, entry) {
			logFilterAction(entry, feed, rule, filterAction)
			return true
		}
	}
	return false
}

func matchesEntryRegexRules(rules string, entry *model.Entry, feed *model.Feed, filterAction filterActionType) bool {
	compiledRegex, err := regexp.Compile(rules)
	if err != nil {
		slog.Warn("Failed on regexp compilation",
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
	rule = strings.TrimSpace(strings.ReplaceAll(rule, "\r\n", ""))
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
		slog.String("filter_action", string(filterAction)),
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

func containsRegexPattern(pattern string, entries []string) bool {
	for _, entry := range entries {
		if matched, _ := regexp.MatchString(pattern, entry); matched {
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
