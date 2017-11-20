// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"github.com/miniflux/miniflux2/server/api/payload"
	"github.com/miniflux/miniflux2/server/core"
)

// CreateUser is the API handler to create a new user.
func (c *Controller) CreateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.Json().Forbidden()
		return
	}

	user, err := payload.DecodeUserPayload(request.GetBody())
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	if err := user.ValidateUserCreation(); err != nil {
		response.Json().BadRequest(err)
		return
	}

	if c.store.UserExists(user.Username) {
		response.Json().BadRequest(errors.New("This user already exists"))
		return
	}

	err = c.store.CreateUser(user)
	if err != nil {
		response.Json().ServerError(errors.New("Unable to create this user"))
		return
	}

	user.Password = ""
	response.Json().Created(user)
}

// UpdateUser is the API handler to update the given user.
func (c *Controller) UpdateUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.Json().Forbidden()
		return
	}

	userID, err := request.GetIntegerParam("userID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	user, err := payload.DecodeUserPayload(request.GetBody())
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	if err := user.ValidateUserModification(); err != nil {
		response.Json().BadRequest(err)
		return
	}

	originalUser, err := c.store.GetUserById(userID)
	if err != nil {
		response.Json().BadRequest(errors.New("Unable to fetch this user from the database"))
		return
	}

	if originalUser == nil {
		response.Json().NotFound(errors.New("User not found"))
		return
	}

	originalUser.Merge(user)
	if err = c.store.UpdateUser(originalUser); err != nil {
		response.Json().ServerError(errors.New("Unable to update this user"))
		return
	}

	response.Json().Created(originalUser)
}

// GetUsers is the API handler to get the list of users.
func (c *Controller) GetUsers(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.Json().Forbidden()
		return
	}

	users, err := c.store.GetUsers()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch the list of users"))
		return
	}

	response.Json().Standard(users)
}

// GetUser is the API handler to fetch the given user.
func (c *Controller) GetUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.Json().Forbidden()
		return
	}

	userID, err := request.GetIntegerParam("userID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	user, err := c.store.GetUserById(userID)
	if err != nil {
		response.Json().BadRequest(errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.Json().NotFound(errors.New("User not found"))
		return
	}

	response.Json().Standard(user)
}

// RemoveUser is the API handler to remove an existing user.
func (c *Controller) RemoveUser(ctx *core.Context, request *core.Request, response *core.Response) {
	if !ctx.IsAdminUser() {
		response.Json().Forbidden()
		return
	}

	userID, err := request.GetIntegerParam("userID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	user, err := c.store.GetUserById(userID)
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.Json().NotFound(errors.New("User not found"))
		return
	}

	if err := c.store.RemoveUser(user.ID); err != nil {
		response.Json().BadRequest(errors.New("Unable to remove this user from the database"))
		return
	}

	response.Json().NoContent()
}
