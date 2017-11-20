// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
	"log"
)

func (c *Controller) ShowUsers(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	users, err := c.store.GetUsers()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("users", args.Merge(tplParams{
		"users": users,
		"menu":  "settings",
	}))
}

func (c *Controller) CreateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("create_user", args.Merge(tplParams{
		"menu": "settings",
		"form": &form.UserForm{},
	}))
}

func (c *Controller) SaveUser(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	userForm := form.NewUserForm(request.GetRequest())
	if err := userForm.ValidateCreation(); err != nil {
		response.Html().Render("create_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	if c.store.UserExists(userForm.Username) {
		response.Html().Render("create_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": "This user already exists.",
		}))
		return
	}

	newUser := userForm.ToUser()
	if err := c.store.CreateUser(newUser); err != nil {
		log.Println(err)
		response.Html().Render("edit_user", args.Merge(tplParams{
			"menu":         "settings",
			"form":         userForm,
			"errorMessage": "Unable to create this user.",
		}))
		return
	}

	response.Redirect(ctx.GetRoute("users"))
}

func (c *Controller) EditUser(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	response.Html().Render("edit_user", args.Merge(tplParams{
		"menu":          "settings",
		"selected_user": selectedUser,
		"form": &form.UserForm{
			Username: selectedUser.Username,
			IsAdmin:  selectedUser.IsAdmin,
		},
	}))
}

func (c *Controller) UpdateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	userForm := form.NewUserForm(request.GetRequest())
	if err := userForm.ValidateModification(); err != nil {
		response.Html().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  err.Error(),
		}))
		return
	}

	if c.store.AnotherUserExists(selectedUser.ID, userForm.Username) {
		response.Html().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  "This user already exists.",
		}))
		return
	}

	userForm.Merge(selectedUser)
	if err := c.store.UpdateUser(selectedUser); err != nil {
		log.Println(err)
		response.Html().Render("edit_user", args.Merge(tplParams{
			"menu":          "settings",
			"selected_user": selectedUser,
			"form":          userForm,
			"errorMessage":  "Unable to update this user.",
		}))
		return
	}

	response.Redirect(ctx.GetRoute("users"))
}

func (c *Controller) RemoveUser(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	if !user.IsAdmin {
		response.Html().Forbidden()
		return
	}

	selectedUser, err := c.getUserFromURL(ctx, request, response)
	if err != nil {
		return
	}

	if err := c.store.RemoveUser(selectedUser.ID); err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Redirect(ctx.GetRoute("users"))
}

func (c *Controller) getUserFromURL(ctx *core.Context, request *core.Request, response *core.Response) (*model.User, error) {
	userID, err := request.GetIntegerParam("userID")
	if err != nil {
		response.Html().BadRequest(err)
		return nil, err
	}

	user, err := c.store.GetUserById(userID)
	if err != nil {
		response.Html().ServerError(err)
		return nil, err
	}

	if user == nil {
		response.Html().NotFound()
		return nil, errors.New("User not found")
	}

	return user, nil
}
