// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

// CurrentUser is the API handler to retrieve the authenticated user.
func (c *Controller) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, user)
}

// CreateUser is the API handler to create a new user.
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	user, err := decodeUserCreationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := user.ValidateUserCreation(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if c.store.UserExists(user.Username) {
		json.BadRequest(w, r, errors.New("This user already exists"))
		return
	}

	err = c.store.CreateUser(user)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	user.Password = ""
	json.Created(w, r, user)
}

// UpdateUser is the API handler to update the given user.
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	userChanges, err := decodeUserModificationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	originalUser, err := c.store.UserByID(userID)
	if err != nil {
		json.BadRequest(w, r, errors.New("Unable to fetch this user from the database"))
		return
	}

	if originalUser == nil {
		json.NotFound(w, r)
		return
	}

	userChanges.Update(originalUser)
	if err := originalUser.ValidateUserModification(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err = c.store.UpdateUser(originalUser); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, originalUser)
}

// Users is the API handler to get the list of users.
func (c *Controller) Users(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	users, err := c.store.Users()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	users.UseTimezone(request.UserTimezone(r))
	json.OK(w, r, users)
}

// UserByID is the API handler to fetch the given user by the ID.
func (c *Controller) UserByID(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := c.store.UserByID(userID)
	if err != nil {
		json.BadRequest(w, r, errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	user.UseTimezone(request.UserTimezone(r))
	json.OK(w, r, user)
}

// UserByUsername is the API handler to fetch the given user by the username.
func (c *Controller) UserByUsername(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	username := request.RouteStringParam(r, "username")
	user, err := c.store.UserByUsername(username)
	if err != nil {
		json.BadRequest(w, r, errors.New("Unable to fetch this user from the database"))
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, user)
}

// RemoveUser is the API handler to remove an existing user.
func (c *Controller) RemoveUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := c.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	if err := c.store.RemoveUser(user.ID); err != nil {
		json.BadRequest(w, r, errors.New("Unable to remove this user from the database"))
		return
	}

	json.NoContent(w, r)
}
