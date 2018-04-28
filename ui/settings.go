// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/ui/form"
)

// ShowSettings shows the settings page.
func (c *Controller) ShowSettings(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	args, err := c.getSettingsFormTemplateArgs(ctx, user, nil)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("settings", ctx.UserLanguage(), args)
}

// UpdateSettings update the settings.
func (c *Controller) UpdateSettings(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	settingsForm := form.NewSettingsForm(request.Request())
	args, err := c.getSettingsFormTemplateArgs(ctx, user, settingsForm)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if err := settingsForm.Validate(); err != nil {
		response.HTML().Render("settings", ctx.UserLanguage(), args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	if c.store.AnotherUserExists(user.ID, settingsForm.Username) {
		response.HTML().Render("settings", ctx.UserLanguage(), args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": "This user already exists.",
		}))
		return
	}

	err = c.store.UpdateUser(settingsForm.Merge(user))
	if err != nil {
		logger.Error("[Controller:UpdateSettings] %v", err)
		response.HTML().Render("settings", ctx.UserLanguage(), args.Merge(tplParams{
			"form":         settingsForm,
			"errorMessage": "Unable to update this user.",
		}))
		return
	}

	ctx.SetFlashMessage(ctx.Translate("Preferences saved!"))
	response.Redirect(ctx.Route("settings"))
}

func (c *Controller) getSettingsFormTemplateArgs(ctx *handler.Context, user *model.User, settingsForm *form.SettingsForm) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return args, err
	}

	if settingsForm == nil {
		args["form"] = form.SettingsForm{
			Username:       user.Username,
			Theme:          user.Theme,
			Language:       user.Language,
			Timezone:       user.Timezone,
			EntryDirection: user.EntryDirection,
		}
	} else {
		args["form"] = settingsForm
	}

	args["menu"] = "settings"
	args["themes"] = model.Themes()
	args["languages"] = locale.AvailableLanguages()
	args["timezones"], err = c.store.Timezones()
	if err != nil {
		return args, err
	}

	return args, nil
}
