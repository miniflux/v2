// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/rewrite"
	"github.com/miniflux/miniflux/reader/sanitizer"
	"github.com/miniflux/miniflux/reader/scraper"
	"github.com/miniflux/miniflux/storage"
)

// FeedProcessor handles the processing of feed contents.
type FeedProcessor struct {
	userID       int64
	store        *storage.Storage
	feed         *model.Feed
	scraperRules string
	rewriteRules string
	crawler      bool
}

// WithCrawler enables the crawler.
func (f *FeedProcessor) WithCrawler(value bool) {
	f.crawler = value
}

// WithScraperRules adds scraper rules to the processing.
func (f *FeedProcessor) WithScraperRules(rules string) {
	f.scraperRules = rules
}

// WithRewriteRules adds rewrite rules to the processing.
func (f *FeedProcessor) WithRewriteRules(rules string) {
	f.rewriteRules = rules
}

// Process applies rewrite and scraper rules.
func (f *FeedProcessor) Process() {
	for _, entry := range f.feed.Entries {
		if f.crawler {
			if f.store.EntryURLExists(f.userID, entry.URL) {
				logger.Debug(`[FeedProcessor] Do not crawl existing entry URL: "%s"`, entry.URL)
			} else {
				content, err := scraper.Fetch(entry.URL, f.scraperRules)
				if err != nil {
					logger.Error("[FeedProcessor] %v", err)
				} else {
					entry.Content = content
				}
			}
		}

		entry.Content = rewrite.Rewriter(entry.URL, entry.Content, f.rewriteRules)
		entry.Content = sanitizer.Sanitize(entry.URL, entry.Content)
	}
}

// NewFeedProcessor returns a new FeedProcessor.
func NewFeedProcessor(userID int64, store *storage.Storage, feed *model.Feed) *FeedProcessor {
	return &FeedProcessor{userID: userID, store: store, feed: feed, crawler: false}
}
