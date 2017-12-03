// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"

	"github.com/miniflux/miniflux2/integration"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
)

// ShowIntegrations renders the page with all external integrations.
func (c *Controller) ShowIntegrations(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	integration, err := c.store.Integration(user.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("integrations", args.Merge(tplParams{
		"menu": "settings",
		"form": form.IntegrationForm{
			PinboardEnabled:      integration.PinboardEnabled,
			PinboardToken:        integration.PinboardToken,
			PinboardTags:         integration.PinboardTags,
			PinboardMarkAsUnread: integration.PinboardMarkAsUnread,
			InstapaperEnabled:    integration.InstapaperEnabled,
			InstapaperUsername:   integration.InstapaperUsername,
			InstapaperPassword:   integration.InstapaperPassword,
		},
	}))
}

// UpdateIntegration updates integration settings.
func (c *Controller) UpdateIntegration(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	integration, err := c.store.Integration(user.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	integrationForm := form.NewIntegrationForm(request.Request())
	integrationForm.Merge(integration)

	err = c.store.UpdateIntegration(integration)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.Redirect(ctx.Route("integrations"))
}

// SaveEntry send the link to external services.
func (c *Controller) SaveEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	user := ctx.LoggedUser()
	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	if entry == nil {
		response.JSON().NotFound(errors.New("Entry not found"))
		return
	}

	settings, err := c.store.Integration(user.ID)
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	go func() {
		integration.SendEntry(entry, settings)
	}()

	response.JSON().Created(map[string]string{"message": "saved"})
}
