// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"fmt"
	"os"
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
)

const (
	largeWeeklyCount = 10080
	noNewTTL         = 0
)

func TestFeedCategorySetter(t *testing.T) {
	feed := &Feed{}
	feed.WithCategoryID(int64(123))

	if feed.Category == nil {
		t.Fatal(`The category field should not be null`)
	}

	if feed.Category.ID != int64(123) {
		t.Error(`The category ID must be set`)
	}
}

func TestFeedErrorCounter(t *testing.T) {
	feed := &Feed{}
	feed.WithTranslatedErrorMessage("Some Error")

	if feed.ParsingErrorMsg != "Some Error" {
		t.Error(`The error message must be set`)
	}

	if feed.ParsingErrorCount != 1 {
		t.Error(`The error counter must be set to 1`)
	}

	feed.ResetErrorCounter()

	if feed.ParsingErrorMsg != "" {
		t.Error(`The error message must be removed`)
	}

	if feed.ParsingErrorCount != 0 {
		t.Error(`The error counter must be set to 0`)
	}
}

func TestFeedCheckedNow(t *testing.T) {
	feed := &Feed{}
	feed.FeedURL = "https://example.org/feed"
	feed.CheckedNow()

	if feed.SiteURL != feed.FeedURL {
		t.Error(`The site URL must not be empty`)
	}

	if feed.CheckedAt.IsZero() {
		t.Error(`The checked date must be set`)
	}
}

func checkTargetInterval(t *testing.T, feed *Feed, targetInterval int, timeBefore time.Time, message string) {
	if feed.NextCheckAt.Before(timeBefore.Add(time.Minute * time.Duration(targetInterval))) {
		t.Errorf(`The next_check_at should be after timeBefore + %s`, message)
	}
	if feed.NextCheckAt.After(time.Now().Add(time.Minute * time.Duration(targetInterval))) {
		t.Errorf(`The next_check_at should be before now + %s`, message)
	}
}

func TestFeedScheduleNextCheckDefault(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	weeklyCount := 10
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := config.Opts.SchedulerRoundRobinMinInterval()
	checkTargetInterval(t, feed, targetInterval, timeBefore, "default SchedulerRoundRobinMinInterval")
}

func TestFeedScheduleNextCheckRoundRobinMinInterval(t *testing.T) {
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "round_robin")
	os.Setenv("SCHEDULER_ROUND_ROBIN_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	weeklyCount := 100
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := minInterval
	checkTargetInterval(t, feed, targetInterval, timeBefore, "round robin min interval")
}

func TestFeedScheduleNextCheckEntryFrequencyMaxInterval(t *testing.T) {
	maxInterval := 5
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", fmt.Sprintf("%d", maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very small weekly count to trigger the max interval
	weeklyCount := 1
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := maxInterval
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency max interval")
}

func TestFeedScheduleNextCheckEntryFrequencyMaxIntervalZeroWeeklyCount(t *testing.T) {
	maxInterval := 5
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", fmt.Sprintf("%d", maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very small weekly count to trigger the max interval
	weeklyCount := 0
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := maxInterval
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency max interval")
}

func TestFeedScheduleNextCheckEntryFrequencyMinInterval(t *testing.T) {
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", fmt.Sprintf("%d", maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := minInterval
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency min interval")
}

func TestFeedScheduleNextCheckEntryFrequencyFactor(t *testing.T) {
	factor := 2
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_FACTOR", fmt.Sprintf("%d", factor))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	weeklyCount := 7
	feed.ScheduleNextCheck(weeklyCount, noNewTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := config.Opts.SchedulerEntryFrequencyMaxInterval() / factor
	checkTargetInterval(t, feed, targetInterval, timeBefore, "factor * count")
}

func TestFeedScheduleNextCheckEntryFrequencySmallNewTTL(t *testing.T) {
	// If the feed has a TTL defined, we use it to make sure we don't check it too often.
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", fmt.Sprintf("%d", maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	// TTL is smaller than minInterval.
	newTTL := minInterval / 2
	feed.ScheduleNextCheck(weeklyCount, newTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := minInterval
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency min interval")

	if feed.NextCheckAt.Before(timeBefore.Add(time.Minute * time.Duration(newTTL))) {
		t.Error(`The next_check_at should be after timeBefore + TTL`)
	}
}

func TestFeedScheduleNextCheckEntryFrequencyLargeNewTTL(t *testing.T) {
	// If the feed has a TTL defined, we use it to make sure we don't check it too often.
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", fmt.Sprintf("%d", maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", fmt.Sprintf("%d", minInterval))

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	// TTL is larger than minInterval.
	newTTL := minInterval * 2
	feed.ScheduleNextCheck(weeklyCount, newTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := newTTL
	checkTargetInterval(t, feed, targetInterval, timeBefore, "TTL")

	if feed.NextCheckAt.Before(timeBefore.Add(time.Minute * time.Duration(minInterval))) {
		t.Error(`The next_check_at should be after timeBefore + entry frequency min interval`)
	}
}
