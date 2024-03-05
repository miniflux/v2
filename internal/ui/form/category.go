// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
)

// CategoryForm represents a feed form in the UI
type CategoryForm struct {
	Title             string
	HideGlobally      string
	Public            string
	ShowOnHomepage    string
	IsHomepageDefault string
}

// NewCategoryForm returns a new CategoryForm.
func NewCategoryForm(r *http.Request) *CategoryForm {
	return &CategoryForm{
		Title:             r.FormValue("title"),
		HideGlobally:      r.FormValue("hide_globally"),
		Public:            r.FormValue("public"),
		ShowOnHomepage:    r.FormValue("show_on_homepage"),
		IsHomepageDefault: r.FormValue("is_homepage_default"),
	}
}
