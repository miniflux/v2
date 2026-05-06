// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"os"
	"strconv"
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
)

const (
	largeWeeklyCount = 10080
	noRefreshDelay   = 0
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

func checkTargetInterval(t *testing.T, feed *Feed, targetInterval time.Duration, timeBefore time.Time, message string) {
	if feed.NextCheckAt.Before(timeBefore.Add(targetInterval)) {
		t.Errorf(`The next_check_at should be after timeBefore + %s`, message)
	}
	if feed.NextCheckAt.After(time.Now().Add(targetInterval)) {
		t.Errorf(`The next_check_at should be before now + %s`, message)
	}
}

func TestFeedScheduleNextCheckRoundRobinDefault(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	feed.ScheduleNextCheck(0, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := config.Opts.SchedulerRoundRobinMinInterval()
	checkTargetInterval(t, feed, targetInterval, timeBefore, "TestFeedScheduleNextCheckRoundRobinDefault")
}

func TestFeedScheduleNextCheckRoundRobinWithRefreshDelayAboveMinInterval(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}

	feed.ScheduleNextCheck(0, config.Opts.SchedulerRoundRobinMinInterval()+30)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	expectedInterval := config.Opts.SchedulerRoundRobinMinInterval() + 30
	checkTargetInterval(t, feed, expectedInterval, timeBefore, "TestFeedScheduleNextCheckRoundRobinWithRefreshDelayAboveMinInterval")
}

func TestFeedScheduleNextCheckRoundRobinWithRefreshDelayBelowMinInterval(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}

	feed.ScheduleNextCheck(0, config.Opts.SchedulerRoundRobinMinInterval()-30)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	expectedInterval := config.Opts.SchedulerRoundRobinMinInterval()
	checkTargetInterval(t, feed, expectedInterval, timeBefore, "TestFeedScheduleNextCheckRoundRobinWithRefreshDelayBelowMinInterval")
}

func TestFeedScheduleNextCheckRoundRobinWithRefreshDelayAboveMaxInterval(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}

	feed.ScheduleNextCheck(0, config.Opts.SchedulerRoundRobinMaxInterval()+30)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	expectedInterval := config.Opts.SchedulerRoundRobinMaxInterval()
	checkTargetInterval(t, feed, expectedInterval, timeBefore, "TestFeedScheduleNextCheckRoundRobinWithRefreshDelayAboveMaxInterval")
}

func TestFeedScheduleNextCheckRoundRobinMinInterval(t *testing.T) {
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "round_robin")
	os.Setenv("SCHEDULER_ROUND_ROBIN_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	feed.ScheduleNextCheck(0, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	expectedInterval := time.Duration(minInterval) * time.Minute
	checkTargetInterval(t, feed, expectedInterval, timeBefore, "TestFeedScheduleNextCheckRoundRobinMinInterval")
}

func TestFeedScheduleNextCheckEntryFrequencyMaxInterval(t *testing.T) {
	maxInterval := 5
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", strconv.Itoa(maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very small weekly count to trigger the max interval
	weeklyCount := 1
	feed.ScheduleNextCheck(weeklyCount, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := time.Duration(maxInterval) * time.Minute
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency max interval")
}

func TestFeedScheduleNextCheckEntryFrequencyMaxIntervalZeroWeeklyCount(t *testing.T) {
	maxInterval := 5
	minInterval := 1
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", strconv.Itoa(maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very small weekly count to trigger the max interval
	weeklyCount := 0
	feed.ScheduleNextCheck(weeklyCount, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := time.Duration(maxInterval) * time.Minute
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency max interval")
}

func TestFeedScheduleNextCheckEntryFrequencyMinInterval(t *testing.T) {
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", strconv.Itoa(maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	feed.ScheduleNextCheck(weeklyCount, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := time.Duration(minInterval) * time.Minute
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency min interval")
}

func TestFeedScheduleNextCheckEntryFrequencyFactor(t *testing.T) {
	factor := 2
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_FACTOR", strconv.Itoa(factor))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	weeklyCount := 7
	feed.ScheduleNextCheck(weeklyCount, noRefreshDelay)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := config.Opts.SchedulerEntryFrequencyMaxInterval() / time.Duration(factor)
	checkTargetInterval(t, feed, targetInterval, timeBefore, "factor * count")
}

func TestFeedScheduleNextCheckEntryFrequencySmallNewTTL(t *testing.T) {
	// If the feed has a TTL defined, we use it to make sure we don't check it too often.
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", strconv.Itoa(maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	// TTL is smaller than minInterval.
	newTTL := time.Duration(minInterval) * time.Minute / 2
	feed.ScheduleNextCheck(weeklyCount, newTTL)

	if feed.NextCheckAt.IsZero() {
		t.Error(`The next_check_at must be set`)
	}

	targetInterval := time.Duration(minInterval) * time.Minute
	checkTargetInterval(t, feed, targetInterval, timeBefore, "entry frequency min interval")

	if feed.NextCheckAt.Before(timeBefore.Add(newTTL)) {
		t.Error(`The next_check_at should be after timeBefore + TTL`)
	}
}

func TestFeedScheduleNextCheckEntryFrequencyLargeNewTTL(t *testing.T) {
	// If the feed has a TTL defined, we use it to make sure we don't check it too often.
	maxInterval := 500
	minInterval := 100
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_frequency")
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", strconv.Itoa(maxInterval))
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", strconv.Itoa(minInterval))

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	timeBefore := time.Now()
	feed := &Feed{}
	// Use a very large weekly count to trigger the min interval
	weeklyCount := largeWeeklyCount
	// TTL is larger than minInterval.
	newTTL := time.Duration(minInterval) * time.Minute * 2
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

func TestFeedScheduleNextCheckRefreshIntervalOverride(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	override := 90
	feed := &Feed{RefreshIntervalMinutes: &override}

	timeBefore := time.Now()
	interval := feed.ScheduleNextCheck(0, noRefreshDelay)

	expected := time.Duration(override) * time.Minute
	if interval != expected {
		t.Errorf(`Expected interval %s, got %s`, expected, interval)
	}

	checkTargetInterval(t, feed, expected, timeBefore, "TestFeedScheduleNextCheckRefreshIntervalOverride")
}

func TestFeedScheduleNextCheckRefreshIntervalOverrideRespectsRefreshDelay(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	// Override is 10 minutes, but the server returns Retry-After 30 minutes:
	// the larger value must win to avoid hammering the publisher.
	override := 10
	feed := &Feed{RefreshIntervalMinutes: &override}
	refreshDelay := 30 * time.Minute

	timeBefore := time.Now()
	interval := feed.ScheduleNextCheck(0, refreshDelay)

	if interval != refreshDelay {
		t.Errorf(`Expected interval %s, got %s`, refreshDelay, interval)
	}

	checkTargetInterval(t, feed, refreshDelay, timeBefore, "TestFeedScheduleNextCheckRefreshIntervalOverrideRespectsRefreshDelay")
}

func TestFeedScheduleNextCheckRefreshIntervalOverrideIgnoresGlobalCap(t *testing.T) {
	os.Clearenv()

	// Round-robin global cap is 1 day by default. Confirm an override above
	// that cap (e.g. 5 days) is honoured rather than clamped, since users
	// who set a per-feed value should get exactly what they asked for.
	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	override := 5 * 24 * 60
	feed := &Feed{RefreshIntervalMinutes: &override}

	interval := feed.ScheduleNextCheck(0, noRefreshDelay)
	expected := time.Duration(override) * time.Minute
	if interval != expected {
		t.Errorf(`Expected interval %s, got %s`, expected, interval)
	}
}

func TestFeedScheduleNextCheckRefreshIntervalNilFallsBackToGlobal(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	feed := &Feed{}
	timeBefore := time.Now()
	interval := feed.ScheduleNextCheck(0, noRefreshDelay)

	if interval != config.Opts.SchedulerRoundRobinMinInterval() {
		t.Errorf(`Expected global round-robin interval %s, got %s`, config.Opts.SchedulerRoundRobinMinInterval(), interval)
	}

	checkTargetInterval(t, feed, interval, timeBefore, "TestFeedScheduleNextCheckRefreshIntervalNilFallsBackToGlobal")
}

func TestFeedScheduleNextCheckRefreshIntervalZeroFallsBackToGlobal(t *testing.T) {
	os.Clearenv()

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	zero := 0
	feed := &Feed{RefreshIntervalMinutes: &zero}
	interval := feed.ScheduleNextCheck(0, noRefreshDelay)

	if interval != config.Opts.SchedulerRoundRobinMinInterval() {
		t.Errorf(`Expected global round-robin interval %s when override is zero, got %s`, config.Opts.SchedulerRoundRobinMinInterval(), interval)
	}
}

func intPtr(value int) *int { return &value }

func TestFeedModificationRequestPatchSetsRefreshInterval(t *testing.T) {
	feed := &Feed{Category: &Category{ID: 1}}
	req := &FeedModificationRequest{RefreshIntervalMinutes: intPtr(60)}
	req.Patch(feed)

	if feed.RefreshIntervalMinutes == nil {
		t.Fatal(`RefreshIntervalMinutes should be set on the feed`)
	}
	if *feed.RefreshIntervalMinutes != 60 {
		t.Errorf(`Expected RefreshIntervalMinutes=60, got %d`, *feed.RefreshIntervalMinutes)
	}
}

func TestFeedModificationRequestPatchClearsRefreshInterval(t *testing.T) {
	existing := 90
	feed := &Feed{Category: &Category{ID: 1}, RefreshIntervalMinutes: &existing}
	req := &FeedModificationRequest{RefreshIntervalMinutes: intPtr(0)}
	req.Patch(feed)

	if feed.RefreshIntervalMinutes != nil {
		t.Errorf(`Expected RefreshIntervalMinutes to be cleared, got %d`, *feed.RefreshIntervalMinutes)
	}
}

func TestFeedModificationRequestPatchLeavesRefreshIntervalAloneWhenNil(t *testing.T) {
	existing := 45
	feed := &Feed{Category: &Category{ID: 1}, RefreshIntervalMinutes: &existing}
	req := &FeedModificationRequest{}
	req.Patch(feed)

	if feed.RefreshIntervalMinutes == nil {
		t.Fatal(`RefreshIntervalMinutes should not be cleared when omitted from the request`)
	}
	if *feed.RefreshIntervalMinutes != 45 {
		t.Errorf(`Expected RefreshIntervalMinutes=45, got %d`, *feed.RefreshIntervalMinutes)
	}
}
