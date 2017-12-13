// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package integration

import (
	"log"

	"github.com/miniflux/miniflux/integration/instapaper"
	"github.com/miniflux/miniflux/integration/pinboard"
	"github.com/miniflux/miniflux/model"
)

// SendEntry send the entry to the activated providers.
func SendEntry(entry *model.Entry, integration *model.Integration) {
	if integration.PinboardEnabled {
		client := pinboard.NewClient(integration.PinboardToken)
		err := client.AddBookmark(entry.URL, entry.Title, integration.PinboardTags, integration.PinboardMarkAsUnread)
		if err != nil {
			log.Println("[Pinboard]", err)
		}
	}

	if integration.InstapaperEnabled {
		client := instapaper.NewClient(integration.InstapaperUsername, integration.InstapaperPassword)
		err := client.AddURL(entry.URL, entry.Title)
		if err != nil {
			log.Println("[Instapaper]", err)
		}
	}
}
