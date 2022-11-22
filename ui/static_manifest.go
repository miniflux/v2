// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/http/route"
	"miniflux.app/model"
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
		Source string `json:"src"`
		Sizes  string `json:"sizes"`
		Type   string `json:"type"`
	}

	type webManifest struct {
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		ShortName       string                 `json:"short_name"`
		StartURL        string                 `json:"start_url"`
		Icons           []webManifestIcon      `json:"icons"`
		ShareTarget     webManifestShareTarget `json:"share_target"`
		Display         string                 `json:"display"`
		ThemeColor      string                 `json:"theme_color"`
		BackgroundColor string                 `json:"background_color"`
	}

	displayMode := "standalone"
	if request.IsAuthenticated(r) {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		displayMode = user.DisplayMode
	}
	themeColor := model.ThemeColor(request.UserTheme(r), "light")
	manifest := &webManifest{
		Name:            "Miniflux",
		ShortName:       "Miniflux",
		Description:     "Minimalist Feed Reader",
		Display:         displayMode,
		StartURL:        route.Path(h.router, "login"),
		ThemeColor:      themeColor,
		BackgroundColor: themeColor,
		Icons: []webManifestIcon{
			{Source: route.Path(h.router, "appIcon", "filename", "icon-120.png"), Sizes: "120x120", Type: "image/png"},
			{Source: route.Path(h.router, "appIcon", "filename", "icon-192.png"), Sizes: "192x192", Type: "image/png"},
			{Source: route.Path(h.router, "appIcon", "filename", "icon-512.png"), Sizes: "512x512", Type: "image/png"},
		},
		ShareTarget: webManifestShareTarget{
			Action:  route.Path(h.router, "bookmarklet"),
			Method:  http.MethodGet,
			Enctype: "application/x-www-form-urlencoded",
			Params:  webManifestShareTargetParams{URL: "uri", Text: "text"},
		},
	}

	json.OK(w, r, manifest)
}
