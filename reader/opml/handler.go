// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"errors"
	"fmt"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/storage"
	"io"
	"log"
)

type OpmlHandler struct {
	store *storage.Storage
}

func (o *OpmlHandler) Export(userID int64) (string, error) {
	feeds, err := o.store.GetFeeds(userID)
	if err != nil {
		log.Println(err)
		return "", errors.New("Unable to fetch feeds.")
	}

	var subscriptions SubcriptionList
	for _, feed := range feeds {
		subscriptions = append(subscriptions, &Subcription{
			Title:        feed.Title,
			FeedURL:      feed.FeedURL,
			SiteURL:      feed.SiteURL,
			CategoryName: feed.Category.Title,
		})
	}

	return Serialize(subscriptions), nil
}

func (o *OpmlHandler) Import(userID int64, data io.Reader) (err error) {
	subscriptions, err := Parse(data)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		if !o.store.FeedURLExists(userID, subscription.FeedURL) {
			var category *model.Category

			if subscription.CategoryName == "" {
				category, err = o.store.GetFirstCategory(userID)
				if err != nil {
					log.Println(err)
					return errors.New("Unable to find first category.")
				}
			} else {
				category, err = o.store.GetCategoryByTitle(userID, subscription.CategoryName)
				if err != nil {
					log.Println(err)
					return errors.New("Unable to search category by title.")
				}

				if category == nil {
					category = &model.Category{
						UserID: userID,
						Title:  subscription.CategoryName,
					}

					err := o.store.CreateCategory(category)
					if err != nil {
						log.Println(err)
						return fmt.Errorf(`Unable to create this category: "%s".`, subscription.CategoryName)
					}
				}
			}

			feed := &model.Feed{
				UserID:   userID,
				Title:    subscription.Title,
				FeedURL:  subscription.FeedURL,
				SiteURL:  subscription.SiteURL,
				Category: category,
			}

			o.store.CreateFeed(feed)
		}
	}

	return nil
}

func NewOpmlHandler(store *storage.Storage) *OpmlHandler {
	return &OpmlHandler{store: store}
}
