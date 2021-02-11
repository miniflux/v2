// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package integration // import "miniflux.app/integration"

import (
	"github.com/gorilla/mux"
	"miniflux.app/config"
	"miniflux.app/http/route"
	"miniflux.app/integration/instapaper"
	"miniflux.app/integration/nunuxkeeper"
	"miniflux.app/integration/pinboard"
	"miniflux.app/integration/pocket"
	"miniflux.app/integration/wallabag"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
)

// SendEntry send the entry to the activated providers.
func SendEntry(entry *model.Entry, integration *model.Integration, router *mux.Router, store *storage.Storage) {
	url := entry.URL
	if entry.Feed.ShareToSave {
		shareCode, err := store.EntryShareCode(entry.UserID, entry.ID)
		if err != nil {
			logger.Error("[Integration] UserID #%d. Error generating share code: %v", integration.UserID, err)
		} else {
			url = route.Path(router, "sharedEntry", "shareCode", shareCode)
		}
	}
	if integration.PinboardEnabled {
		client := pinboard.NewClient(integration.PinboardToken)
		err := client.AddBookmark(
			url,
			entry.Title,
			integration.PinboardTags,
			integration.PinboardMarkAsUnread,
		)

		if err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.InstapaperEnabled {
		client := instapaper.NewClient(integration.InstapaperUsername, integration.InstapaperPassword)
		if err := client.AddURL(url, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.WallabagEnabled {
		client := wallabag.NewClient(
			integration.WallabagURL,
			integration.WallabagClientID,
			integration.WallabagClientSecret,
			integration.WallabagUsername,
			integration.WallabagPassword,
		)

		if err := client.AddEntry(url, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.NunuxKeeperEnabled {
		client := nunuxkeeper.NewClient(
			integration.NunuxKeeperURL,
			integration.NunuxKeeperAPIKey,
		)

		if err := client.AddEntry(url, entry.Title, entry.Content); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.PocketEnabled {
		client := pocket.NewClient(config.Opts.PocketConsumerKey(integration.PocketConsumerKey), integration.PocketAccessToken)
		if err := client.AddURL(url, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}
}
