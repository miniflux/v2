// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/version"
)

// AboutPage shows the about page.
func (c *Controller) AboutPage(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("about", args.Merge(tplParams{
		"version":    version.Version,
		"build_date": version.BuildDate,
		"menu":       "settings",
	}))
}
