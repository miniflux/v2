// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/static"
)

// AppIcon renders application icons.
func (c *Controller) AppIcon(w http.ResponseWriter, r *http.Request) {
	filename := request.Param(r, "filename", "favicon.png")
	encodedBlob, found := static.Binaries[filename]
	if !found {
		logger.Info("[Controller:AppIcon] This icon doesn't exists: %s", filename)
		html.NotFound(w)
		return
	}

	blob, err := base64.StdEncoding.DecodeString(encodedBlob)
	if err != nil {
		logger.Error("[Controller:AppIcon] %v", err)
		html.NotFound(w)
		return
	}

	response.Cache(w, r, "image/png", static.BinariesChecksums[filename], blob, 48*time.Hour)
}
