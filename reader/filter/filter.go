// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter

import (
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/rewrite"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/reader/scraper"
	"miniflux.app/storage"
)

// Apply executes all entry filters.
func Apply(store *storage.Storage, feed *model.Feed) {
	for _, entry := range feed.Entries {
		if feed.Crawler {
			if !store.EntryURLExists(feed.UserID, entry.URL) {
				content, err := scraper.Fetch(entry.URL, feed.ScraperRules, feed.UserAgent)
				if err != nil {
					logger.Error("Unable to crawl this entry: %q => %v", entry.URL, err)
				} else {
					// We replace the entry content only if the scraper doesn't return any error.
					entry.Content = content
				}
			}
		}

		entry.Content = rewrite.Rewriter(entry.URL, entry.Content, feed.RewriteRules)

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
		entry.Content = sanitizer.Sanitize(entry.URL, entry.Content)
	}
}
