// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
)

func (h *handler) showWebManifest(w http.ResponseWriter, r *http.Request) {
	type webManifestShareTargetParams struct {
		URL  string `json:"url"`
		Text string `json:"text"`
	}

	type webManifestShareTarget struct {
		Action  string                       `json:"action"`
		Method  string                       `json:"method"`
		Enctype string                       `json:"enctype"`
		Params  webManifestShareTargetParams `json:"params"`
	}

	type webManifestIcon struct {
		Source  string `json:"src"`
		Sizes   string `json:"sizes,omitempty"`
		Type    string `json:"type,omitempty"`
		Purpose string `json:"purpose,omitempty"`
	}

	type webManifestShortcut struct {
		Name  string            `json:"name"`
		URL   string            `json:"url"`
		Icons []webManifestIcon `json:"icons,omitempty"`
	}

	type webManifest struct {
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		ShortName       string                 `json:"short_name"`
		StartURL        string                 `json:"start_url"`
		Icons           []webManifestIcon      `json:"icons"`
		ShareTarget     webManifestShareTarget `json:"share_target"`
		Display         string                 `json:"display"`
		BackgroundColor string                 `json:"background_color"`
		Shortcuts       []webManifestShortcut  `json:"shortcuts"`
	}

	displayMode := "standalone"
	labelNewFeed := "Add Feed"
	labelUnreadMenu := "Unread"
	labelStarredMenu := "Starred"
	labelHistoryMenu := "History"
	labelFeedsMenu := "Feeds"
	labelCategoriesMenu := "Categories"
	labelSearchMenu := "Search"
	labelSettingsMenu := "Settings"
	if request.IsAuthenticated(r) {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		displayMode = user.DisplayMode
		printer := locale.NewPrinter(user.Language)
		labelNewFeed = printer.Print("menu.add_feed")
		labelUnreadMenu = printer.Print("menu.unread")
		labelStarredMenu = printer.Print("menu.starred")
		labelHistoryMenu = printer.Print("menu.history")
		labelFeedsMenu = printer.Print("menu.feeds")
		labelCategoriesMenu = printer.Print("menu.categories")
		labelSearchMenu = printer.Print("menu.search")
		labelSettingsMenu = printer.Print("menu.settings")
	}
	themeColor := model.ThemeColor(request.UserTheme(r), "light")
	manifest := &webManifest{
		Name:            "Miniflux",
		ShortName:       "Miniflux",
		Description:     "Minimalist Feed Reader",
		Display:         displayMode,
		StartURL:        h.routePath("/"),
		BackgroundColor: themeColor,
		Icons: []webManifestIcon{
			{Source: h.iconPath("icon-120.png"), Sizes: "120x120", Type: "image/png", Purpose: "any"},
			{Source: h.iconPath("icon-192.png"), Sizes: "192x192", Type: "image/png", Purpose: "any"},
			{Source: h.iconPath("icon-512.png"), Sizes: "512x512", Type: "image/png", Purpose: "any"},
			{Source: h.iconPath("maskable-icon-120.png"), Sizes: "120x120", Type: "image/png", Purpose: "maskable"},
			{Source: h.iconPath("maskable-icon-192.png"), Sizes: "192x192", Type: "image/png", Purpose: "maskable"},
			{Source: h.iconPath("maskable-icon-512.png"), Sizes: "512x512", Type: "image/png", Purpose: "maskable"},
		},
		ShareTarget: webManifestShareTarget{
			Action:  h.routePath("/bookmarklet"),
			Method:  http.MethodGet,
			Enctype: "application/x-www-form-urlencoded",
			Params:  webManifestShareTargetParams{URL: "uri", Text: "text"},
		},
		Shortcuts: []webManifestShortcut{
			{Name: labelNewFeed, URL: h.routePath("/subscribe"), Icons: []webManifestIcon{{Source: h.iconPath("add-feed-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelUnreadMenu, URL: h.routePath("/unread"), Icons: []webManifestIcon{{Source: h.iconPath("unread-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelStarredMenu, URL: h.routePath("/starred"), Icons: []webManifestIcon{{Source: h.iconPath("starred-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelHistoryMenu, URL: h.routePath("/history"), Icons: []webManifestIcon{{Source: h.iconPath("history-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelFeedsMenu, URL: h.routePath("/feeds"), Icons: []webManifestIcon{{Source: h.iconPath("feeds-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelCategoriesMenu, URL: h.routePath("/categories"), Icons: []webManifestIcon{{Source: h.iconPath("categories-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelSearchMenu, URL: h.routePath("/search"), Icons: []webManifestIcon{{Source: h.iconPath("search-icon.png"), Sizes: "240x240", Type: "image/png"}}},
			{Name: labelSettingsMenu, URL: h.routePath("/settings"), Icons: []webManifestIcon{{Source: h.iconPath("settings-icon.png"), Sizes: "240x240", Type: "image/png"}}},
		},
	}

	response.JSON(w, r, manifest)
}
