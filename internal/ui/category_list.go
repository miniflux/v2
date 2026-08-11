// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showCategoryListPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	categories, err := h.store.CategoriesWithFeedCount(user.ID, user.CategoriesSortingOrder)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	items := make([]itemListItem, 0, len(categories))
	for _, category := range categories {
		feedCount := intPtrValue(category.FeedCount)
		unreadCount := intPtrValue(category.TotalUnread)
		actions := []itemListAction{
			{
				Class:    "item-meta-icons-entries",
				URL:      h.routePath("/category/%d/entries", category.ID),
				Icon:     "entries",
				LabelKey: "page.categories.entries",
			},
			{
				Class:    "item-meta-icons-feeds",
				URL:      h.routePath("/category/%d/feeds", category.ID),
				Icon:     "feeds",
				LabelKey: "page.categories.feeds",
			},
			{
				Class:    "item-meta-icons-edit",
				URL:      h.routePath("/category/%d/edit", category.ID),
				Icon:     "edit",
				LabelKey: "menu.edit_category",
			},
		}

		if feedCount == 0 {
			actions = append(actions, itemListAction{
				Class:    "item-meta-icons-delete",
				URL:      h.routePath("/category/%d/remove", category.ID),
				Icon:     "delete",
				LabelKey: "action.remove",
				Confirm:  true,
			})
		}

		if unreadCount > 0 {
			actions = append(actions, itemListAction{
				Class:    "item-meta-icons-mark-as-read",
				URL:      h.routePath("/category/%d/mark-all-as-read", category.ID),
				Icon:     "read",
				LabelKey: "menu.mark_all_as_read",
				Confirm:  true,
			})
		}

		items = append(items, itemListItem{
			ID:            "category-" + strconv.FormatInt(category.ID, 10),
			Title:         category.Title,
			TitleURL:      h.routePath("/category/%d/entries", category.ID),
			UnreadCount:   unreadCount,
			MetaCount:     feedCount,
			MetaPluralKey: "page.categories.feed_count",
			MetaEmptyKey:  "page.categories.no_feed",
			Actions:       actions,
		})
	}

	view := view.New(h.tpl, r)
	view.Set("categories", categories)
	view.Set("items", items)
	view.Set("total", len(categories))
	view.Set("menu", "categories")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)

	response.HTML(w, r, view.Render("categories"))
}
