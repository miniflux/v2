// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) currentUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, user)
}

func (h *handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		response.JSONForbidden(w, r)
		return
	}

	var userCreationRequest model.UserCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&userCreationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateUserCreationWithPassword(h.store, &userCreationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	user, err := h.store.CreateUser(&userCreationRequest)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, user)
}

func (h *handler) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")
	if userID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid user ID"))
		return
	}

	var userModificationRequest model.UserModificationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&userModificationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	originalUser, err := h.store.UserByID(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if originalUser == nil {
		response.JSONNotFound(w, r)
		return
	}

	if !request.IsAdminUser(r) {
		if originalUser.ID != request.UserID(r) {
			response.JSONForbidden(w, r)
			return
		}

		if userModificationRequest.IsAdmin != nil && *userModificationRequest.IsAdmin {
			response.JSONBadRequest(w, r, errors.New("only administrators can change permissions of standard users"))
			return
		}
	}

	if validationErr := validator.ValidateUserModification(h.store, originalUser.ID, &userModificationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	userModificationRequest.Patch(originalUser)
	if err = h.store.UpdateUser(originalUser); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, originalUser)
}

func (h *handler) markUserAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")
	if userID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid user ID"))
		return
	}

	if userID != request.UserID(r) {
		response.JSONForbidden(w, r)
		return
	}

	if _, err := h.store.UserByID(userID); err != nil {
		response.JSONNotFound(w, r)
		return
	}

	if err := h.store.MarkAllAsRead(userID); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) getIntegrationsStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	if _, err := h.store.UserByID(userID); err != nil {
		response.JSONNotFound(w, r)
		return
	}

	hasIntegrations := h.store.HasSaveEntry(userID)

	response.JSON(w, r, integrationsStatusResponse{HasIntegrations: hasIntegrations})
}

func (h *handler) usersHandler(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		response.JSONForbidden(w, r)
		return
	}

	users, err := h.store.Users()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	users.UseTimezone(request.UserTimezone(r))
	response.JSON(w, r, users)
}

func (h *handler) dispatchUserLookupHandler(w http.ResponseWriter, r *http.Request) {
	identifier := request.RouteStringParam(r, "identifier")
	userID := request.RouteInt64Param(r, "identifier")
	if userID > 0 {
		r.SetPathValue("userID", identifier)
		h.userByIDHandler(w, r)
		return
	}

	r.SetPathValue("username", identifier)
	h.userByUsernameHandler(w, r)
}

func (h *handler) userByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		response.JSONForbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	if userID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid user ID"))
		return
	}

	user, err := h.store.UserByID(userID)
	if err != nil {
		response.JSONBadRequest(w, r, errors.New("unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	user.UseTimezone(request.UserTimezone(r))
	response.JSON(w, r, user)
}

func (h *handler) userByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		response.JSONForbidden(w, r)
		return
	}

	username := request.RouteStringParam(r, "username")
	user, err := h.store.UserByUsername(username)
	if err != nil {
		response.JSONBadRequest(w, r, errors.New("unable to fetch this user from the database"))
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	response.JSON(w, r, user)
}

func (h *handler) removeUserHandler(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		response.JSONForbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	if userID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid user ID"))
		return
	}

	user, err := h.store.UserByID(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	if user.ID == request.UserID(r) {
		response.JSONBadRequest(w, r, errors.New("you cannot remove yourself"))
		return
	}

	h.store.RemoveUserAsync(user.ID)
	response.NoContent(w, r)
}
