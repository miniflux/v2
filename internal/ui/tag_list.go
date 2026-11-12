// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/url"
	"strconv"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showTagListPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	tags, err := h.store.TagsWithEntryCount(user.ID, user.CategoriesSortingOrder)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	items := make([]itemListItem, 0, len(tags))
	for i, tag := range tags {
		tagEntriesURL := h.routePath("/tags/%s/entries/all", url.PathEscape(tag.Title))
		actions := []itemListAction{
			{
				Class:    "item-meta-icons-entries",
				URL:      tagEntriesURL,
				Icon:     "entries",
				LabelKey: "page.categories.entries",
			},
		}

		if tag.TotalUnread > 0 {
			actions = append(actions, itemListAction{
				Class:    "item-meta-icons-mark-as-read",
				URL:      h.routePath("/tags/%s/mark-all-as-read", url.PathEscape(tag.Title)),
				Icon:     "read",
				LabelKey: "menu.mark_all_as_read",
				Confirm:  true,
			})
		}

		items = append(items, itemListItem{
			ID:            "tag-" + strconv.Itoa(i),
			Title:         tag.Title,
			TitleURL:      tagEntriesURL,
			UnreadCount:   tag.TotalUnread,
			MetaCount:     tag.TotalEntries,
			MetaPluralKey: "page.total_entry_count",
			Actions:       actions,
		})
	}

	view := view.New(h.tpl, r)
	view.Set("items", items)
	view.Set("total", len(tags))
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)

	response.HTML(w, r, view.Render("tags"))
}
