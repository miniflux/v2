// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/locale"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
	"log"
)

func (c *Controller) ShowSettings(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	args, err := c.getSettingsFormTemplateArgs(ctx, user, nil)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("settings", args)
}

func (c *Controller) UpdateSettings(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	settingsForm := form.NewSettingsForm(request.Request())
	args, err := c.getSettingsFormTemplateArgs(ctx, user, settingsForm)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if err := settingsForm.Validate(); err != nil {
		response.HTML().Render("settings", args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	if c.store.AnotherUserExists(user.ID, settingsForm.Username) {
		response.HTML().Render("settings", args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": "This user already exists.",
		}))
		return
	}

	err = c.store.UpdateUser(settingsForm.Merge(user))
	if err != nil {
		log.Println(err)
		response.HTML().Render("settings", args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": "Unable to update this user.",
		}))
		return
	}

	response.Redirect(ctx.GetRoute("settings"))
}

func (c *Controller) getSettingsFormTemplateArgs(ctx *core.Context, user *model.User, settingsForm *form.SettingsForm) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return args, err
	}

	if settingsForm == nil {
		args["form"] = form.SettingsForm{
			Username: user.Username,
			Theme:    user.Theme,
			Language: user.Language,
			Timezone: user.Timezone,
		}
	} else {
		args["form"] = settingsForm
	}

	args["menu"] = "settings"
	args["themes"] = model.GetThemes()
	args["languages"] = locale.GetAvailableLanguages()
	args["timezones"], err = c.store.GetTimezones()
	if err != nil {
		return args, err
	}

	return args, nil
}
