// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package subscription

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/miniflux/miniflux2/errors"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/reader/http"
	"github.com/miniflux/miniflux2/reader/url"

	"github.com/PuerkitoBio/goquery"
)

var (
	errConnectionFailure = "Unable to open this link: %v"
	errUnreadableDoc     = "Unable to analyze this page: %v"
)

// FindSubscriptions downloads and try to find one or more subscriptions from an URL.
func FindSubscriptions(websiteURL string) (Subscriptions, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[FindSubscriptions] url=%s", websiteURL))

	client := http.NewClient(websiteURL)
	response, err := client.Get()
	if err != nil {
		return nil, errors.NewLocalizedError(errConnectionFailure, err)
	}

	var buffer bytes.Buffer
	io.Copy(&buffer, response.Body)
	reader := bytes.NewReader(buffer.Bytes())

	if format := feed.DetectFeedFormat(reader); format != feed.FormatUnknown {
		var subscriptions Subscriptions
		subscriptions = append(subscriptions, &Subscription{
			Title: response.EffectiveURL,
			URL:   response.EffectiveURL,
			Type:  format,
		})

		return subscriptions, nil
	}

	reader.Seek(0, io.SeekStart)
	return parseDocument(response.EffectiveURL, bytes.NewReader(buffer.Bytes()))
}

func parseDocument(websiteURL string, data io.Reader) (Subscriptions, error) {
	var subscriptions Subscriptions
	queries := map[string]string{
		"link[type='application/rss+xml']":  "rss",
		"link[type='application/atom+xml']": "atom",
		"link[type='application/json']":     "json",
	}

	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return nil, errors.NewLocalizedError(errUnreadableDoc, err)
	}

	for query, kind := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			subscription := new(Subscription)
			subscription.Type = kind

			if title, exists := s.Attr("title"); exists {
				subscription.Title = title
			} else {
				subscription.Title = "Feed"
			}

			if feedURL, exists := s.Attr("href"); exists {
				subscription.URL, _ = url.AbsoluteURL(websiteURL, feedURL)
			}

			if subscription.Title == "" {
				subscription.Title = subscription.URL
			}

			if subscription.URL != "" {
				log.Println("[FindSubscriptions]", subscription)
				subscriptions = append(subscriptions, subscription)
			}
		})
	}

	return subscriptions, nil
}
