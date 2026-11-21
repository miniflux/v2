// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if !loggedUser.IsAdmin {
		response.HTMLForbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	selectedUser, err := h.store.UserByID(userID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if selectedUser == nil {
		response.HTMLNotFound(w, r)
		return
	}

	// Changing your own password must go through the authentication page,
	// which asks for the current password.
	if selectedUser.ID == loggedUser.ID {
		response.HTMLBadRequest(w, r, errors.New("you cannot edit yourself"))
		return
	}

	userForm := form.NewUserForm(r)

	view := view.New(h.tpl, r)
	view.Set("menu", "settings")
	view.Set("user", loggedUser)
	navMetadata, _ := h.store.GetNavMetadata(loggedUser.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("selected_user", selectedUser)
	view.Set("form", userForm)

	if validationErr := userForm.ValidateModification(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		response.HTML(w, r, view.Render("edit_user"))
		return
	}

	userModificationRequest := &model.UserModificationRequest{
		Username: model.OptionalString(userForm.Username),
	}

	if validationErr := validator.ValidateUserModification(h.store, selectedUser.ID, userModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		response.HTML(w, r, view.Render("edit_user"))
		return
	}

	userForm.Merge(selectedUser)
	if err := h.store.UpdateUser(selectedUser); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if userForm.Password != "" {
		if err := h.store.RemoveUserWebSessionsExcept(selectedUser.ID, ""); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	response.HTMLRedirect(w, r, h.routePath("/users"))
}
