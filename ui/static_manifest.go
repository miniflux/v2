// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/response/json"
	"github.com/miniflux/miniflux/http/route"
)

// WebManifest renders web manifest file.
func (c *Controller) WebManifest(w http.ResponseWriter, r *http.Request) {
	type webManifestIcon struct {
		Source string `json:"src"`
		Sizes  string `json:"sizes"`
		Type   string `json:"type"`
	}

	type webManifest struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		ShortName   string            `json:"short_name"`
		StartURL    string            `json:"start_url"`
		Icons       []webManifestIcon `json:"icons"`
		Display     string            `json:"display"`
	}

	manifest := &webManifest{
		Name:        "Miniflux",
		ShortName:   "Miniflux",
		Description: "Minimalist Feed Reader",
		Display:     "minimal-ui",
		StartURL:    route.Path(c.router, "unread"),
		Icons: []webManifestIcon{
			webManifestIcon{Source: route.Path(c.router, "appIcon", "filename", "icon-120.png"), Sizes: "120x120", Type: "image/png"},
			webManifestIcon{Source: route.Path(c.router, "appIcon", "filename", "icon-192.png"), Sizes: "192x192", Type: "image/png"},
			webManifestIcon{Source: route.Path(c.router, "appIcon", "filename", "icon-512.png"), Sizes: "512x512", Type: "image/png"},
		},
	}

	json.OK(w, manifest)
}
