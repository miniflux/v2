// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/static"
)

// Favicon renders the application favicon.
func (c *Controller) Favicon(w http.ResponseWriter, r *http.Request) {
	blob, err := base64.StdEncoding.DecodeString(static.Binaries["favicon.ico"])
	if err != nil {
		logger.Error("[Controller:Favicon] %v", err)
		html.NotFound(w)
		return
	}

	response.Cache(w, r, "image/x-icon", static.BinariesChecksums["favicon.ico"], blob, 48*time.Hour)
}
