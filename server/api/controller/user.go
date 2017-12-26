// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"

	"github.com/miniflux/miniflux/server/api/payload"
	"github.com/miniflux/miniflux/server/core"
)

// CreateUser is the API handler to create a new user.
func (c *Controller) CreateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	user, err := payload.DecodeUserPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if err := user.ValidateUserCreation(); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if c.store.UserExists(user.Username) {
		response.JSON().BadRequest(errors.New("This user already exists"))
		return
	}

	err = c.store.CreateUser(user)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to create this user"))
		return
	}

	user.Password = ""
	response.JSON().Created(user)
}

// UpdateUser is the API handler to update the given user.
func (c *Controller) UpdateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	userID, err := request.IntegerParam("userID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	user, err := payload.DecodeUserPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if err := user.ValidateUserModification(); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	originalUser, err := c.store.UserByID(userID)
	if err != nil {
		response.JSON().BadRequest(errors.New("Unable to fetch this user from the database"))
		return
	}

	if originalUser == nil {
		response.JSON().NotFound(errors.New("User not found"))
		return
	}

	originalUser.Merge(user)
	if err = c.store.UpdateUser(originalUser); err != nil {
		response.JSON().ServerError(errors.New("Unable to update this user"))
		return
	}

	response.JSON().Created(originalUser)
}

// Users is the API handler to get the list of users.
func (c *Controller) Users(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	users, err := c.store.Users()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch the list of users"))
		return
	}

	response.JSON().Standard(users)
}

// UserByID is the API handler to fetch the given user by the ID.
func (c *Controller) UserByID(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	userID, err := request.IntegerParam("userID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	user, err := c.store.UserByID(userID)
	if err != nil {
		response.JSON().BadRequest(errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.JSON().NotFound(errors.New("User not found"))
		return
	}

	response.JSON().Standard(user)
}

// UserByUsername is the API handler to fetch the given user by the username.
func (c *Controller) UserByUsername(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	username := request.StringParam("username", "")
	user, err := c.store.UserByUsername(username)
	if err != nil {
		response.JSON().BadRequest(errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.JSON().NotFound(errors.New("User not found"))
		return
	}

	response.JSON().Standard(user)
}

// RemoveUser is the API handler to remove an existing user.
func (c *Controller) RemoveUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.JSON().Forbidden()
		return
	}

	userID, err := request.IntegerParam("userID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	user, err := c.store.UserByID(userID)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.JSON().NotFound(errors.New("User not found"))
		return
	}

	if err := c.store.RemoveUser(user.ID); err != nil {
		response.JSON().BadRequest(errors.New("Unable to remove this user from the database"))
		return
	}

	response.JSON().NoContent()
}
