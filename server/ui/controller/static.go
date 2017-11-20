// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"encoding/base64"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/static"
	"log"
	"time"
)

func (c *Controller) Stylesheet(ctx *core.Context, request *core.Request, response *core.Response) {
	stylesheet := request.GetStringParam("name", "white")
	body := static.Stylesheets["common"]
	etag := static.StylesheetsChecksums["common"]

	if theme, found := static.Stylesheets[stylesheet]; found {
		body += theme
		etag += static.StylesheetsChecksums[stylesheet]
	}

	response.Cache("text/css", etag, []byte(body), 48*time.Hour)
}

func (c *Controller) Javascript(ctx *core.Context, request *core.Request, response *core.Response) {
	response.Cache("text/javascript", static.JavascriptChecksums["app"], []byte(static.Javascript["app"]), 48*time.Hour)
}

func (c *Controller) Favicon(ctx *core.Context, request *core.Request, response *core.Response) {
	blob, err := base64.StdEncoding.DecodeString(static.Binaries["favicon.ico"])
	if err != nil {
		log.Println(err)
		response.Html().NotFound()
		return
	}

	response.Cache("image/x-icon", static.BinariesChecksums["favicon.ico"], blob, 48*time.Hour)
}
