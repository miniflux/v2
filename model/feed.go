// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"fmt"
	"math"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/http/client"
)

// Feed represents a feed in the application.
type Feed struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	FeedURL            string    `json:"feed_url"`
	SiteURL            string    `json:"site_url"`
	Title              string    `json:"title"`
	CheckedAt          time.Time `json:"checked_at"`
	NextCheckAt        time.Time `json:"next_check_at"`
	EtagHeader         string    `json:"etag_header"`
	LastModifiedHeader string    `json:"last_modified_header"`
	ParsingErrorMsg    string    `json:"parsing_error_message"`
	ParsingErrorCount  int       `json:"parsing_error_count"`
	ScraperRules       string    `json:"scraper_rules"`
	RewriteRules       string    `json:"rewrite_rules"`
	Crawler            bool      `json:"crawler"`
	UserAgent          string    `json:"user_agent"`
	Username           string    `json:"username"`
	Password           string    `json:"password"`
	Disabled           bool      `json:"disabled"`
	Category           *Category `json:"category,omitempty"`
	Entries            Entries   `json:"entries,omitempty"`
	Icon               *FeedIcon `json:"icon"`
	UnreadCount        int       `json:"-"`
	ReadCount          int       `json:"-"`
}

const (
	// SchedulerEntryCountBased represnets the name of the scheduler based on entry counts.
	SchedulerEntryCountBased = "entry_count_based"
)

func (f *Feed) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, FeedURL=%s, SiteURL=%s, Title=%s, Category={%s}",
		f.ID,
		f.UserID,
		f.FeedURL,
		f.SiteURL,
		f.Title,
		f.Category,
	)
}

// WithClientResponse updates feed attributes from an HTTP request.
func (f *Feed) WithClientResponse(response *client.Response) {
	f.EtagHeader = response.ETag
	f.LastModifiedHeader = response.LastModified
	f.FeedURL = response.EffectiveURL
}

// WithCategoryID initializes the category attribute of the feed.
func (f *Feed) WithCategoryID(categoryID int64) {
	f.Category = &Category{ID: categoryID}
}

// WithBrowsingParameters defines browsing parameters.
func (f *Feed) WithBrowsingParameters(crawler bool, userAgent, username, password, scraperRules, rewriteRules string) {
	f.Crawler = crawler
	f.UserAgent = userAgent
	f.Username = username
	f.Password = password
	f.ScraperRules = scraperRules
	f.RewriteRules = rewriteRules
}

// WithError adds a new error message and increment the error counter.
func (f *Feed) WithError(message string) {
	f.ParsingErrorCount++
	f.ParsingErrorMsg = message
}

// ResetErrorCounter removes all previous errors.
func (f *Feed) ResetErrorCounter() {
	f.ParsingErrorCount = 0
	f.ParsingErrorMsg = ""
}

// CheckedNow set attribute values when the feed is refreshed.
func (f *Feed) CheckedNow() {
	f.CheckedAt = time.Now()

	if f.SiteURL == "" {
		f.SiteURL = f.FeedURL
	}
}

// ScheduleNextCheck set "next_check_at" of a feed based on the scheduler selected from the configuration.
func (f *Feed) ScheduleNextCheck(weeklyCount int) {
	var nextCheckAt time.Time
	switch strings.ToLower(config.Opts.PollingScheduler()) {
	case SchedulerEntryCountBased:
		var intervalMinutes int
		if weeklyCount == 0 {
			intervalMinutes = config.Opts.SchedulerCountBasedMaxInterval()
		} else {
			intervalMinutes = int(math.Round(float64(7*24*60) / float64(weeklyCount)))
		}
		intervalMinutes = int(math.Min(float64(intervalMinutes), float64(config.Opts.SchedulerCountBasedMaxInterval())))
		intervalMinutes = int(math.Max(float64(intervalMinutes), float64(config.Opts.SchedulerCountBasedMinInterval())))
		nextCheckAt = time.Now().Add(time.Minute * time.Duration(intervalMinutes))
	default:
		// round robin
		// omit the interval because they are same for all feeds.
		nextCheckAt = time.Now()
	}
	f.NextCheckAt = nextCheckAt
}

// Feeds is a list of feed
type Feeds []*Feed
