// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"

	"miniflux.app/errors"
	"miniflux.app/model"
)

// CategoryForm represents a feed form in the UI
type CategoryForm struct {
	Title string
}

// Validate makes sure the form values are valid.
func (c CategoryForm) Validate() error {
	if c.Title == "" {
		return errors.NewLocalizedError("The title is mandatory.")
	}
	return nil
}

// Merge update the given category fields.
func (c CategoryForm) Merge(category *model.Category) *model.Category {
	category.Title = c.Title
	return category
}

// NewCategoryForm returns a new CategoryForm.
func NewCategoryForm(r *http.Request) *CategoryForm {
	return &CategoryForm{
		Title: r.FormValue("title"),
	}
}
