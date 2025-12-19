// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"

	// Integration clients for isolated tests
	"miniflux.app/v2/internal/integration/apprise"
	"miniflux.app/v2/internal/integration/betula"
	"miniflux.app/v2/internal/integration/cubox"
	"miniflux.app/v2/internal/integration/discord"
	"miniflux.app/v2/internal/integration/espial"
	"miniflux.app/v2/internal/integration/instapaper"
	"miniflux.app/v2/internal/integration/karakeep"
	"miniflux.app/v2/internal/integration/linkace"
	"miniflux.app/v2/internal/integration/linkding"
	"miniflux.app/v2/internal/integration/linktaco"
	"miniflux.app/v2/internal/integration/linkwarden"
	"miniflux.app/v2/internal/integration/matrixbot"
	"miniflux.app/v2/internal/integration/notion"
	"miniflux.app/v2/internal/integration/ntfy"
	"miniflux.app/v2/internal/integration/nunuxkeeper"
	"miniflux.app/v2/internal/integration/omnivore"
	"miniflux.app/v2/internal/integration/pinboard"
	"miniflux.app/v2/internal/integration/pushover"
	"miniflux.app/v2/internal/integration/raindrop"
	"miniflux.app/v2/internal/integration/readeck"
	"miniflux.app/v2/internal/integration/readwise"
	"miniflux.app/v2/internal/integration/rssbridge"
	"miniflux.app/v2/internal/integration/shaarli"
	"miniflux.app/v2/internal/integration/shiori"
	"miniflux.app/v2/internal/integration/slack"
	"miniflux.app/v2/internal/integration/telegrambot"
	"miniflux.app/v2/internal/integration/wallabag"
	"miniflux.app/v2/internal/integration/webhook"
)

// testerFunc defines a per-integration test sender.
type testerFunc func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error

// testers holds isolated test functions per integration.
var testers = map[string]testerFunc{
	"readeck": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := readeck.NewClient(cfg.ReadeckURL, cfg.ReadeckAPIKey, cfg.ReadeckLabels, cfg.ReadeckOnlyURL)
		return client.CreateBookmark(entry.URL, entry.Title, entry.Content)
	},
	"linkding": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := linkding.NewClient(cfg.LinkdingURL, cfg.LinkdingAPIKey, cfg.LinkdingTags, cfg.LinkdingMarkAsUnread)
		return client.CreateBookmark(entry.URL, entry.Title)
	},
	"linkace": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := linkace.NewClient(cfg.LinkAceURL, cfg.LinkAceAPIKey, cfg.LinkAceTags, cfg.LinkAcePrivate, cfg.LinkAceCheckDisabled)
		return client.AddURL(entry.URL, entry.Title)
	},
	"pinboard": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := pinboard.NewClient(cfg.PinboardToken)
		return client.CreateBookmark(entry.URL, entry.Title, cfg.PinboardTags, cfg.PinboardMarkAsUnread)
	},
	"instapaper": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := instapaper.NewClient(cfg.InstapaperUsername, cfg.InstapaperPassword)
		return client.AddURL(entry.URL, entry.Title)
	},
	"wallabag": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := wallabag.NewClient(cfg.WallabagURL, cfg.WallabagClientID, cfg.WallabagClientSecret, cfg.WallabagUsername, cfg.WallabagPassword, cfg.WallabagTags, cfg.WallabagOnlyURL)
		return client.CreateEntry(entry.URL, entry.Title, entry.Content)
	},
	"notion": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := notion.NewClient(cfg.NotionToken, cfg.NotionPageID)
		return client.UpdateDocument(entry.URL, entry.Title)
	},
	"nunuxkeeper": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := nunuxkeeper.NewClient(cfg.NunuxKeeperURL, cfg.NunuxKeeperAPIKey)
		return client.AddEntry(entry.URL, entry.Title, entry.Content)
	},
	"espial": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := espial.NewClient(cfg.EspialURL, cfg.EspialAPIKey)
		return client.CreateLink(entry.URL, entry.Title, cfg.EspialTags)
	},
	"linktaco": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := linktaco.NewClient(cfg.LinktacoAPIToken, cfg.LinktacoOrgSlug, cfg.LinktacoTags, cfg.LinktacoVisibility)
		return client.CreateBookmark(entry.URL, entry.Title, entry.Content)
	},
	"linkwarden": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := linkwarden.NewClient(cfg.LinkwardenURL, cfg.LinkwardenAPIKey)
		return client.CreateBookmark(entry.URL, entry.Title)
	},
	"readwise": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := readwise.NewClient(cfg.ReadwiseAPIKey)
		return client.CreateDocument(entry.URL)
	},
	"cubox": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := cubox.NewClient(cfg.CuboxAPILink)
		return client.SaveLink(entry.URL)
	},
	"shiori": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := shiori.NewClient(cfg.ShioriURL, cfg.ShioriUsername, cfg.ShioriPassword)
		return client.CreateBookmark(entry.URL, entry.Title)
	},
	"shaarli": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := shaarli.NewClient(cfg.ShaarliURL, cfg.ShaarliAPISecret)
		return client.CreateLink(entry.URL, entry.Title)
	},
	"webhook": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := webhook.NewClient(cfg.WebhookURL, cfg.WebhookSecret)
		return client.SendNewEntriesWebhookEvent(feed, model.Entries{entry})
	},
	"omnivore": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := omnivore.NewClient(cfg.OmnivoreAPIKey, cfg.OmnivoreURL)
		return client.SaveUrl(entry.URL)
	},
	"karakeep": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := karakeep.NewClient(cfg.KarakeepAPIKey, cfg.KarakeepURL, cfg.KarakeepTags)
		return client.SaveURL(entry.URL)
	},
	"raindrop": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := raindrop.NewClient(cfg.RaindropToken, cfg.RaindropCollectionID, cfg.RaindropTags)
		return client.CreateRaindrop(entry.URL, entry.Title)
	},
	"matrix": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		return matrixbot.PushEntries(feed, model.Entries{entry}, cfg.MatrixBotURL, cfg.MatrixBotUser, cfg.MatrixBotPassword, cfg.MatrixBotChatID)
	},
	"apprise": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := apprise.NewClient(cfg.AppriseServicesURL, cfg.AppriseURL)
		return client.SendNotification(feed, model.Entries{entry})
	},
	"discord": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := discord.NewClient(cfg.DiscordWebhookLink)
		return client.SendDiscordMsg(feed, model.Entries{entry})
	},
	"slack": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := slack.NewClient(cfg.SlackWebhookLink)
		return client.SendSlackMsg(feed, model.Entries{entry})
	},
	"pushover": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := pushover.New(cfg.PushoverUser, cfg.PushoverToken, feed.PushoverPriority, cfg.PushoverDevice, cfg.PushoverPrefix)
		return client.SendMessages(feed, model.Entries{entry})
	},
	"ntfy": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		client := ntfy.NewClient(cfg.NtfyURL, cfg.NtfyTopic, cfg.NtfyAPIToken, cfg.NtfyUsername, cfg.NtfyPassword, cfg.NtfyIconURL, cfg.NtfyInternalLinks, feed.NtfyPriority)
		return client.SendMessages(feed, model.Entries{entry})
	},
	"telegram": func(cfg *model.Integration, feed *model.Feed, entry *model.Entry) error {
		return telegrambot.PushEntry(feed, entry, cfg.TelegramBotToken, cfg.TelegramBotChatID, cfg.TelegramBotTopicID, cfg.TelegramBotDisableWebPagePreview, cfg.TelegramBotDisableNotification, cfg.TelegramBotDisableButtons)
	},
	"betula": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		client := betula.NewClient(cfg.BetulaURL, cfg.BetulaToken)
		return client.CreateBookmark(entry.URL, entry.Title, entry.Tags)
	},
	"fever": func(cfg *model.Integration, _ *model.Feed, _ *model.Entry) error {
		return testFeverIntegration(cfg)
	},
	"googlereader": func(cfg *model.Integration, _ *model.Feed, _ *model.Entry) error {
		return testGoogleReaderIntegration(cfg)
	},
	"rssbridge": func(cfg *model.Integration, _ *model.Feed, entry *model.Entry) error {
		return testRSSBridgeIntegration(cfg, entry)
	},
}

// testIntegration sends a synthetic entry to a selected integration using the saved settings.
func (h *handler) testIntegration(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))
	userID := request.UserID(r)

	target := strings.ToLower(strings.TrimSpace(r.FormValue("target")))
	if target == "" {
		html.BadRequest(w, r, errors.New("missing integration target"))
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Build a synthetic entry/feed for testing.
	testEntry := model.NewEntry()
	testEntry.Title = "Miniflux Integration Test"
	testEntry.URL = "https://miniflux.app/"
	testEntry.Content = "This is a test entry generated by Miniflux to verify your integration setup."

	testFeed := &model.Feed{Title: "Miniflux Test Feed", SiteURL: "https://miniflux.app/"}
	testEntry.Feed = testFeed

	tester, ok := testers[target]
	if !ok {
		html.BadRequest(w, r, errors.New("unknown integration target"))
		return
	}

	if err := tester(userIntegrations, testFeed, testEntry); err != nil {
		sess.NewFlashErrorMessage(printer.Printf("%s test failed: %v", target, err))
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	sess.NewFlashMessage(printer.Printf("Sent test entry to %s", target))
	html.Redirect(w, r, route.Path(h.router, "integrations"))
}

func testFeverIntegration(cfg *model.Integration) error {
	if cfg.FeverToken == "" {
		return errors.New("missing Fever API token")
	}

	baseURL := strings.TrimSuffix(config.Opts.BaseURL(), "/")
	endpoint := baseURL + "/fever/"
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	query.Set("api_key", cfg.FeverToken)
	request.URL.RawQuery = query.Encode()

	client := &http.Client{Timeout: config.Opts.HTTPClientTimeout()}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", response.StatusCode)
	}

	var result struct {
		Authenticated int `json:"auth"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if result.Authenticated != 1 {
		return errors.New("authentication failed")
	}

	return nil
}

func testGoogleReaderIntegration(cfg *model.Integration) error {
	if cfg.GoogleReaderUsername == "" || cfg.GoogleReaderPassword == "" {
		return errors.New("missing Google Reader credentials")
	}

	baseURL := strings.TrimSuffix(config.Opts.BaseURL(), "/")
	endpoint := baseURL + "/accounts/ClientLogin"
	data := url.Values{}
	data.Set("Email", cfg.GoogleReaderUsername)
	data.Set("Passwd", cfg.GoogleReaderPassword)
	data.Set("output", "json")

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: config.Opts.HTTPClientTimeout()}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", response.StatusCode)
	}

	var result struct {
		Auth string `json:"Auth"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if result.Auth == "" {
		return errors.New("missing auth token")
	}

	return nil
}

func testRSSBridgeIntegration(cfg *model.Integration, entry *model.Entry) error {
	if cfg.RSSBridgeURL == "" {
		return errors.New("missing RSS-Bridge URL")
	}

	_, err := rssbridge.DetectBridges(cfg.RSSBridgeURL, cfg.RSSBridgeToken, entry.URL)
	return err
}
