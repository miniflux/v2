// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
)

func (h *handler) currentUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, user)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userCreationRequest, err := decodeUserCreationRequest(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	user := model.NewUser()
	user.Username = userCreationRequest.Username
	user.Password = userCreationRequest.Password
	user.IsAdmin = userCreationRequest.IsAdmin
	user.GoogleID = userCreationRequest.GoogleID
	user.OpenIDConnectID = userCreationRequest.OpenIDConnectID

	if err := user.ValidateUserCreation(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if h.store.UserExists(user.Username) {
		json.BadRequest(w, r, errors.New("This user already exists"))
		return
	}

	err = h.store.CreateUser(user)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	user.Password = ""
	json.Created(w, r, user)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")
	userChanges, err := decodeUserModificationRequest(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	originalUser, err := h.store.UserByID(userID)
	if err != nil {
		json.BadRequest(w, r, errors.New("Unable to fetch this user from the database"))
		return
	}

	if originalUser == nil {
		json.NotFound(w, r)
		return
	}

	if !request.IsAdminUser(r) {
		if originalUser.ID != request.UserID(r) {
			json.Forbidden(w, r)
			return
		}

		if userChanges.IsAdmin != nil && *userChanges.IsAdmin {
			json.BadRequest(w, r, errors.New("Only administrators can change permissions of standard users"))
			return
		}
	}

	userChanges.Update(originalUser)
	if err := originalUser.ValidateUserModification(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err = h.store.UpdateUser(originalUser); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, originalUser)
}

func (h *handler) markUserAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")
	if userID != request.UserID(r) {
		json.Forbidden(w, r)
		return
	}

	if _, err := h.store.UserByID(userID); err != nil {
		json.NotFound(w, r)
		return
	}

	if err := h.store.MarkAllAsRead(userID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) users(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	users, err := h.store.Users()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	users.UseTimezone(request.UserTimezone(r))
	json.OK(w, r, users)
}

func (h *handler) userByID(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := h.store.UserByID(userID)
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

func (h *handler) userByUsername(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	username := request.RouteStringParam(r, "username")
	user, err := h.store.UserByUsername(username)
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

func (h *handler) removeUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	if user.ID == request.UserID(r) {
		json.BadRequest(w, r, errors.New("You cannot remove yourself"))
		return
	}

	h.store.RemoveUserAsync(user.ID)
	json.NoContent(w, r)
}
