// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/base64"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/logger"
	"miniflux.app/ui/static"
)

// AppIcon renders application icons.
func (c *Controller) AppIcon(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "filename")
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
