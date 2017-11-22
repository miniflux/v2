// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"log"

	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
)

// ShowCategories shows the page with all categories.
func (c *Controller) ShowCategories(ctx *core.Context, request *core.Request, response *core.Response) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	user := ctx.LoggedUser()
	categories, err := c.store.GetCategoriesWithFeedCount(user.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("categories", args.Merge(tplParams{
		"categories": categories,
		"total":      len(categories),
		"menu":       "categories",
	}))
}

// ShowCategoryEntries shows all entries for the given category.
func (c *Controller) ShowCategoryEntries(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	offset := request.QueryIntegerParam("offset", 0)

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	category, err := c.getCategoryFromURL(ctx, request, response)
	if err != nil {
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(category.ID)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.DefaultSortingDirection)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOffset(offset)
	builder.WithLimit(NbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("category_entries", args.Merge(tplParams{
		"category":   category,
		"entries":    entries,
		"total":      count,
		"pagination": c.getPagination(ctx.Route("categoryEntries", "categoryID", category.ID), count, offset),
		"menu":       "categories",
	}))
}

// CreateCategory shows the form to create a new category.
func (c *Controller) CreateCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("create_category", args.Merge(tplParams{
		"menu": "categories",
	}))
}

// SaveCategory validate and save the new category into the database.
func (c *Controller) SaveCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	categoryForm := form.NewCategoryForm(request.Request())
	if err := categoryForm.Validate(); err != nil {
		response.HTML().Render("create_category", args.Merge(tplParams{
			"errorMessage": err.Error(),
		}))
		return
	}

	duplicateCategory, err := c.store.GetCategoryByTitle(user.ID, categoryForm.Title)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if duplicateCategory != nil {
		response.HTML().Render("create_category", args.Merge(tplParams{
			"errorMessage": "This category already exists.",
		}))
		return
	}

	category := model.Category{Title: categoryForm.Title, UserID: user.ID}
	err = c.store.CreateCategory(&category)
	if err != nil {
		log.Println(err)
		response.HTML().Render("create_category", args.Merge(tplParams{
			"errorMessage": "Unable to create this category.",
		}))
		return
	}

	response.Redirect(ctx.Route("categories"))
}

// EditCategory shows the form to modify a category.
func (c *Controller) EditCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	category, err := c.getCategoryFromURL(ctx, request, response)
	if err != nil {
		log.Println(err)
		return
	}

	args, err := c.getCategoryFormTemplateArgs(ctx, user, category, nil)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("edit_category", args)
}

// UpdateCategory validate and update a category.
func (c *Controller) UpdateCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	category, err := c.getCategoryFromURL(ctx, request, response)
	if err != nil {
		log.Println(err)
		return
	}

	categoryForm := form.NewCategoryForm(request.Request())
	args, err := c.getCategoryFormTemplateArgs(ctx, user, category, categoryForm)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if err := categoryForm.Validate(); err != nil {
		response.HTML().Render("edit_category", args.Merge(tplParams{
			"errorMessage": err.Error(),
		}))
		return
	}

	if c.store.AnotherCategoryExists(user.ID, category.ID, categoryForm.Title) {
		response.HTML().Render("edit_category", args.Merge(tplParams{
			"errorMessage": "This category already exists.",
		}))
		return
	}

	err = c.store.UpdateCategory(categoryForm.Merge(category))
	if err != nil {
		log.Println(err)
		response.HTML().Render("edit_category", args.Merge(tplParams{
			"errorMessage": "Unable to update this category.",
		}))
		return
	}

	response.Redirect(ctx.Route("categories"))
}

// RemoveCategory delete a category from the database.
func (c *Controller) RemoveCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	category, err := c.getCategoryFromURL(ctx, request, response)
	if err != nil {
		return
	}

	if err := c.store.RemoveCategory(user.ID, category.ID); err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.Redirect(ctx.Route("categories"))
}

func (c *Controller) getCategoryFromURL(ctx *core.Context, request *core.Request, response *core.Response) (*model.Category, error) {
	categoryID, err := request.IntegerParam("categoryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return nil, err
	}

	user := ctx.LoggedUser()
	category, err := c.store.GetCategory(user.ID, categoryID)
	if err != nil {
		response.HTML().ServerError(err)
		return nil, err
	}

	if category == nil {
		response.HTML().NotFound()
		return nil, errors.New("Category not found")
	}

	return category, nil
}

func (c *Controller) getCategoryFormTemplateArgs(ctx *core.Context, user *model.User, category *model.Category, categoryForm *form.CategoryForm) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return nil, err
	}

	if categoryForm == nil {
		args["form"] = form.CategoryForm{
			Title: category.Title,
		}
	} else {
		args["form"] = categoryForm
	}

	args["category"] = category
	args["menu"] = "categories"
	return args, nil
}
