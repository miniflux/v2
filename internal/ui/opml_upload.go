// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/opml"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) uploadOPML(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		slog.Error("OPML file upload error",
			slog.Int64("user_id", user.ID),
			slog.Any("error", err),
		)

		response.HTMLRedirect(w, r, h.routePath("/import"))
		return
	}
	defer file.Close()

	slog.Info("OPML file uploaded",
		slog.Int64("user_id", user.ID),
		slog.String("file_name", fileHeader.Filename),
		slog.Int64("file_size", fileHeader.Size),
	)

	view := view.New(h.tpl, r)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	if fileHeader.Size == 0 {
		view.Set("errorMessage", locale.NewLocalizedError("error.empty_file").Translate(user.Language))
		response.HTML(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, file); impErr != nil {
		view.Set("errorMessage", impErr)
		response.HTML(w, r, view.Render("import"))
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/feeds"))
}

func (h *handler) fetchOPML(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	opmlFileURL := strings.TrimSpace(r.FormValue("url"))
	if opmlFileURL == "" {
		response.HTMLRedirect(w, r, h.routePath("/import"))
		return
	}

	slog.Info("Fetching OPML file remotely",
		slog.Int64("user_id", user.ID),
		slog.String("opml_file_url", opmlFileURL),
	)

	view := view.New(h.tpl, r)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(opmlFileURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch OPML file", slog.String("opml_file_url", opmlFileURL), slog.Any("error", localizedError.Error()))
		view.Set("errorMessage", localizedError.Translate(user.Language))
		response.HTML(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, responseHandler.Body(config.Opts.HTTPClientMaxBodySize())); impErr != nil {
		view.Set("errorMessage", impErr)
		response.HTML(w, r, view.Render("import"))
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/feeds"))
}
