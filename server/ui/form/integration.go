// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"net/http"

	"github.com/miniflux/miniflux/model"
)

// IntegrationForm represents user integration settings form.
type IntegrationForm struct {
	PinboardEnabled      bool
	PinboardToken        string
	PinboardTags         string
	PinboardMarkAsUnread bool
	InstapaperEnabled    bool
	InstapaperUsername   string
	InstapaperPassword   string
	FeverEnabled         bool
	FeverUsername        string
	FeverPassword        string
}

// Merge copy form values to the model.
func (i IntegrationForm) Merge(integration *model.Integration) {
	integration.PinboardEnabled = i.PinboardEnabled
	integration.PinboardToken = i.PinboardToken
	integration.PinboardTags = i.PinboardTags
	integration.PinboardMarkAsUnread = i.PinboardMarkAsUnread
	integration.InstapaperEnabled = i.InstapaperEnabled
	integration.InstapaperUsername = i.InstapaperUsername
	integration.InstapaperPassword = i.InstapaperPassword
	integration.FeverEnabled = i.FeverEnabled
	integration.FeverUsername = i.FeverUsername
	integration.FeverPassword = i.FeverPassword
}

// NewIntegrationForm returns a new AuthForm.
func NewIntegrationForm(r *http.Request) *IntegrationForm {
	return &IntegrationForm{
		PinboardEnabled:      r.FormValue("pinboard_enabled") == "1",
		PinboardToken:        r.FormValue("pinboard_token"),
		PinboardTags:         r.FormValue("pinboard_tags"),
		PinboardMarkAsUnread: r.FormValue("pinboard_mark_as_unread") == "1",
		InstapaperEnabled:    r.FormValue("instapaper_enabled") == "1",
		InstapaperUsername:   r.FormValue("instapaper_username"),
		InstapaperPassword:   r.FormValue("instapaper_password"),
		FeverEnabled:         r.FormValue("fever_enabled") == "1",
		FeverUsername:        r.FormValue("fever_username"),
		FeverPassword:        r.FormValue("fever_password"),
	}
}
