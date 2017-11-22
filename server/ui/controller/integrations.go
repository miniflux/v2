// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import "github.com/miniflux/miniflux2/server/core"

// ShowIntegrations renders the page with all external integrations.
func (c *Controller) ShowIntegrations(ctx *core.Context, request *core.Request, response *core.Response) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("integrations", args.Merge(tplParams{
		"menu": "settings",
	}))
}
