// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"encoding/base64"
	"time"

	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/server/core"
	"github.com/miniflux/miniflux/server/static"
)

// Stylesheet renders the CSS.
func (c *Controller) Stylesheet(ctx *core.Context, request *core.Request, response *core.Response) {
	stylesheet := request.StringParam("name", "white")
	body := static.Stylesheets["common"]
	etag := static.StylesheetsChecksums["common"]

	if theme, found := static.Stylesheets[stylesheet]; found {
		body += theme
		etag += static.StylesheetsChecksums[stylesheet]
	}

	response.Cache("text/css; charset=utf-8", etag, []byte(body), 48*time.Hour)
}

// Javascript renders application client side code.
func (c *Controller) Javascript(ctx *core.Context, request *core.Request, response *core.Response) {
	response.Cache("text/javascript; charset=utf-8", static.JavascriptChecksums["app"], []byte(static.Javascript["app"]), 48*time.Hour)
}

// Favicon renders the application favicon.
func (c *Controller) Favicon(ctx *core.Context, request *core.Request, response *core.Response) {
	blob, err := base64.StdEncoding.DecodeString(static.Binaries["favicon.ico"])
	if err != nil {
		logger.Error("[Controller:Favicon] %v", err)
		response.HTML().NotFound()
		return
	}

	response.Cache("image/x-icon", static.BinariesChecksums["favicon.ico"], blob, 48*time.Hour)
}

// AppIcon returns application icons.
func (c *Controller) AppIcon(ctx *core.Context, request *core.Request, response *core.Response) {
	filename := request.StringParam("filename", "favicon.png")
	encodedBlob, found := static.Binaries[filename]
	if !found {
		logger.Info("[Controller:AppIcon] This icon doesn't exists: %s", filename)
		response.HTML().NotFound()
		return
	}

	blob, err := base64.StdEncoding.DecodeString(encodedBlob)
	if err != nil {
		logger.Error("[Controller:AppIcon] %v", err)
		response.HTML().NotFound()
		return
	}

	response.Cache("image/png", static.BinariesChecksums[filename], blob, 48*time.Hour)
}
