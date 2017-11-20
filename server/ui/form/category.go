// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"errors"
	"github.com/miniflux/miniflux2/model"
	"net/http"
)

// CategoryForm represents a feed form in the UI
type CategoryForm struct {
	Title string
}

func (c CategoryForm) Validate() error {
	if c.Title == "" {
		return errors.New("The title is mandatory.")
	}
	return nil
}

func (c CategoryForm) Merge(category *model.Category) *model.Category {
	category.Title = c.Title
	return category
}

func NewCategoryForm(r *http.Request) *CategoryForm {
	return &CategoryForm{
		Title: r.FormValue("title"),
	}
}
