// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/rewrite"
	"github.com/miniflux/miniflux2/reader/sanitizer"
)

// FeedProcessor handles the processing of feed contents.
type FeedProcessor struct {
	feed         *model.Feed
	scraperRules string
	rewriteRules string
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
		entry.Content = sanitizer.Sanitize(entry.URL, entry.Content)
		entry.Content = rewrite.Rewriter(entry.URL, entry.Content, f.rewriteRules)
	}
}

// NewFeedProcessor returns a new FeedProcessor.
func NewFeedProcessor(feed *model.Feed) *FeedProcessor {
	return &FeedProcessor{feed: feed}
}
