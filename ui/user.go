// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"errors"

	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/ui/form"
)

// ShowUsers shows the list of users.
func (c *Controller) ShowUsers(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	users, err := c.store.Users()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("users", args.Merge(tplParams{
		"users": users,
		"menu":  "settings",
	}))
}

// CreateUser shows the user creation form.
func (c *Controller) CreateUser(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("create_user", args.Merge(tplParams{
		"menu": "settings",
		"form": &form.UserForm{},
	}))
}

// SaveUser validate and save the new user into the database.
func (c *Controller) SaveUser(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	userForm := form.NewUserForm(request.Request())
	if err := userForm.ValidateCreation(); err != nil {
		response.HTML().Render("create_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	if c.store.UserExists(userForm.Username) {
		response.HTML().Render("create_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": "This user already exists.",
		}))
		return
	}

	newUser := userForm.ToUser()
	if err := c.store.CreateUser(newUser); err != nil {
		logger.Error("[Controller:SaveUser] %v", err)
		response.HTML().Render("edit_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": "Unable to create this user.",
		}))
		return
	}

	response.Redirect(ctx.Route("users"))
}

// EditUser shows the form to edit a user.
func (c *Controller) EditUser(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	response.HTML().Render("edit_user", args.Merge(tplParams{
		"menu":          "settings",
		"selected_user": selectedUser,
		"form": &form.UserForm{
			Username: selectedUser.Username,
			IsAdmin:  selectedUser.IsAdmin,
		},
	}))
}

// UpdateUser validate and update a user.
func (c *Controller) UpdateUser(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	userForm := form.NewUserForm(request.Request())
	if err := userForm.ValidateModification(); err != nil {
		response.HTML().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  err.Error(),
		}))
		return
	}

	if c.store.AnotherUserExists(selectedUser.ID, userForm.Username) {
		response.HTML().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  "This user already exists.",
		}))
		return
	}

	userForm.Merge(selectedUser)
	if err := c.store.UpdateUser(selectedUser); err != nil {
		logger.Error("[Controller:UpdateUser] %v", err)
		response.HTML().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  "Unable to update this user.",
		}))
		return
	}

	response.Redirect(ctx.Route("users"))
}

// RemoveUser deletes a user from the database.
func (c *Controller) RemoveUser(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()
	if !user.IsAdmin {
		response.HTML().Forbidden()
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	if err := c.store.RemoveUser(selectedUser.ID); err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.Redirect(ctx.Route("users"))
}

func (c *Controller) getUserFromURL(ctx *handler.Context, request *handler.Request, response *handler.Response) (*model.User, error) {
	userID, err := request.IntegerParam("userID")
	if err != nil {
		response.HTML().BadRequest(err)
		return nil, err
	}

	user, err := c.store.UserByID(userID)
	if err != nil {
		response.HTML().ServerError(err)
		return nil, err
	}

	if user == nil {
		response.HTML().NotFound()
		return nil, errors.New("User not found")
	}

	return user, nil
}
