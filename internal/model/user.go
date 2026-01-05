// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/timezone"
	"miniflux.app/v2/internal/version"
)

// MarkReadBehavior list all possible behaviors for automatically marking an entry as read
type MarkReadBehavior string

const (
	NoAutoMarkAsRead                           MarkReadBehavior = "no-auto"
	MarkAsReadOnView                           MarkReadBehavior = "on-view"
	MarkAsReadOnViewButWaitForPlayerCompletion MarkReadBehavior = "on-view-but-wait-for-player-completion"
	MarkAsReadOnlyOnPlayerCompletion           MarkReadBehavior = "on-player-completion"
)

// User represents a user in the system.
type User struct {
	ID                              int64      `json:"id"`
	Username                        string     `json:"username"`
	Password                        string     `json:"-"`
	IsAdmin                         bool       `json:"is_admin"`
	Theme                           string     `json:"theme"`
	Language                        string     `json:"language"`
	Timezone                        string     `json:"timezone"`
	EntryDirection                  string     `json:"entry_sorting_direction"`
	EntryOrder                      string     `json:"entry_sorting_order"`
	Stylesheet                      string     `json:"stylesheet"`
	CustomJS                        string     `json:"custom_js"`
	ExternalFontHosts               string     `json:"external_font_hosts"`
	GoogleID                        string     `json:"google_id"`
	OpenIDConnectID                 string     `json:"openid_connect_id"`
	EntriesPerPage                  int        `json:"entries_per_page"`
	KeyboardShortcuts               bool       `json:"keyboard_shortcuts"`
	ShowReadingTime                 bool       `json:"show_reading_time"`
	EntrySwipe                      bool       `json:"entry_swipe"`
	GestureNav                      string     `json:"gesture_nav"`
	LastLoginAt                     *time.Time `json:"last_login_at"`
	DisplayMode                     string     `json:"display_mode"`
	DefaultReadingSpeed             int        `json:"default_reading_speed"`
	CJKReadingSpeed                 int        `json:"cjk_reading_speed"`
	DefaultHomePage                 string     `json:"default_home_page"`
	CategoriesSortingOrder          string     `json:"categories_sorting_order"`
	MarkReadOnView                  bool       `json:"mark_read_on_view"`
	MarkReadOnMediaPlayerCompletion bool       `json:"mark_read_on_media_player_completion"`
	MediaPlaybackRate               float64    `json:"media_playback_rate"`
	BlockFilterEntryRules           string     `json:"block_filter_entry_rules"`
	KeepFilterEntryRules            string     `json:"keep_filter_entry_rules"`
	AlwaysOpenExternalLinks         bool       `json:"always_open_external_links"`
	OpenExternalLinksInNewTab       bool       `json:"open_external_links_in_new_tab"`
}

// UserCreationRequest represents the request to create a user.
type UserCreationRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	GoogleID        string `json:"google_id"`
	OpenIDConnectID string `json:"openid_connect_id"`
}

// UserModificationRequest represents the request to update a user.
type UserModificationRequest struct {
	Username                        *string  `json:"username"`
	Password                        *string  `json:"password"`
	Theme                           *string  `json:"theme"`
	Language                        *string  `json:"language"`
	Timezone                        *string  `json:"timezone"`
	EntryDirection                  *string  `json:"entry_sorting_direction"`
	EntryOrder                      *string  `json:"entry_sorting_order"`
	Stylesheet                      *string  `json:"stylesheet"`
	CustomJS                        *string  `json:"custom_js"`
	ExternalFontHosts               *string  `json:"external_font_hosts"`
	GoogleID                        *string  `json:"google_id"`
	OpenIDConnectID                 *string  `json:"openid_connect_id"`
	EntriesPerPage                  *int     `json:"entries_per_page"`
	IsAdmin                         *bool    `json:"is_admin"`
	KeyboardShortcuts               *bool    `json:"keyboard_shortcuts"`
	ShowReadingTime                 *bool    `json:"show_reading_time"`
	EntrySwipe                      *bool    `json:"entry_swipe"`
	GestureNav                      *string  `json:"gesture_nav"`
	DisplayMode                     *string  `json:"display_mode"`
	DefaultReadingSpeed             *int     `json:"default_reading_speed"`
	CJKReadingSpeed                 *int     `json:"cjk_reading_speed"`
	DefaultHomePage                 *string  `json:"default_home_page"`
	CategoriesSortingOrder          *string  `json:"categories_sorting_order"`
	MarkReadOnView                  *bool    `json:"mark_read_on_view"`
	MarkReadOnMediaPlayerCompletion *bool    `json:"mark_read_on_media_player_completion"`
	MediaPlaybackRate               *float64 `json:"media_playback_rate"`
	BlockFilterEntryRules           *string  `json:"block_filter_entry_rules"`
	KeepFilterEntryRules            *string  `json:"keep_filter_entry_rules"`
	AlwaysOpenExternalLinks         *bool    `json:"always_open_external_links"`
	OpenExternalLinksInNewTab       *bool    `json:"open_external_links_in_new_tab"`
}

// Patch updates the User object with the modification request.
func (u *UserModificationRequest) Patch(user *User) {
	if u.Username != nil && *u.Username != "" && !config.Opts.DisableLocalAuth() {
		user.Username = *u.Username
	}

	if u.Password != nil && *u.Password != "" {
		user.Password = *u.Password
	}

	if u.IsAdmin != nil {
		user.IsAdmin = *u.IsAdmin
	}

	if u.Theme != nil && *u.Theme != "" {
		user.Theme = *u.Theme
	}

	if u.Language != nil && *u.Language != "" {
		user.Language = *u.Language
	}

	if u.Timezone != nil && *u.Timezone != "" {
		user.Timezone = *u.Timezone
	}

	if u.EntryDirection != nil {
		user.EntryDirection = *u.EntryDirection
	}

	if u.EntryOrder != nil {
		user.EntryOrder = *u.EntryOrder
	}

	if u.Stylesheet != nil {
		user.Stylesheet = *u.Stylesheet
	}

	if u.CustomJS != nil {
		user.CustomJS = *u.CustomJS
	}

	if u.ExternalFontHosts != nil {
		user.ExternalFontHosts = *u.ExternalFontHosts
	}

	if u.GoogleID != nil {
		user.GoogleID = *u.GoogleID
	}

	if u.OpenIDConnectID != nil {
		user.OpenIDConnectID = *u.OpenIDConnectID
	}

	if u.EntriesPerPage != nil {
		user.EntriesPerPage = *u.EntriesPerPage
	}

	if u.KeyboardShortcuts != nil {
		user.KeyboardShortcuts = *u.KeyboardShortcuts
	}

	if u.ShowReadingTime != nil {
		user.ShowReadingTime = *u.ShowReadingTime
	}

	if u.EntrySwipe != nil {
		user.EntrySwipe = *u.EntrySwipe
	}

	if u.GestureNav != nil {
		user.GestureNav = *u.GestureNav
	}

	if u.DisplayMode != nil {
		user.DisplayMode = *u.DisplayMode
	}

	if u.DefaultReadingSpeed != nil && *u.DefaultReadingSpeed != 0 {
		user.DefaultReadingSpeed = *u.DefaultReadingSpeed
	}

	if u.CJKReadingSpeed != nil && *u.CJKReadingSpeed != 0 {
		user.CJKReadingSpeed = *u.CJKReadingSpeed
	}

	if u.DefaultHomePage != nil {
		user.DefaultHomePage = *u.DefaultHomePage
	}

	if u.CategoriesSortingOrder != nil {
		user.CategoriesSortingOrder = *u.CategoriesSortingOrder
	}

	if u.MarkReadOnView != nil {
		user.MarkReadOnView = *u.MarkReadOnView
	}

	if u.MarkReadOnMediaPlayerCompletion != nil {
		user.MarkReadOnMediaPlayerCompletion = *u.MarkReadOnMediaPlayerCompletion
	}

	if u.MediaPlaybackRate != nil {
		user.MediaPlaybackRate = *u.MediaPlaybackRate
	}

	if u.BlockFilterEntryRules != nil {
		user.BlockFilterEntryRules = *u.BlockFilterEntryRules
	}

	if u.KeepFilterEntryRules != nil {
		user.KeepFilterEntryRules = *u.KeepFilterEntryRules
	}

	if u.AlwaysOpenExternalLinks != nil {
		user.AlwaysOpenExternalLinks = *u.AlwaysOpenExternalLinks
	}

	if u.OpenExternalLinksInNewTab != nil {
		user.OpenExternalLinksInNewTab = *u.OpenExternalLinksInNewTab
	}
}

// UseTimezone converts last login date to the given timezone.
func (u *User) UseTimezone(tz string) {
	if u.LastLoginAt != nil {
		*u.LastLoginAt = timezone.Convert(tz, *u.LastLoginAt)
	}
}

// Users represents a list of users.
type Users []*User

// UseTimezone converts last login timestamp of all users to the given timezone.
func (u Users) UseTimezone(tz string) {
	for _, user := range u {
		user.UseTimezone(tz)
	}
}

// MarkAsReadBehavior returns the MarkReadBehavior from the given MarkReadOnView and MarkReadOnMediaPlayerCompletion values.
// Useful to convert the values from the User model to the form
func MarkAsReadBehavior(markReadOnView, markReadOnMediaPlayerCompletion bool) MarkReadBehavior {
	switch {
	case markReadOnView && !markReadOnMediaPlayerCompletion:
		return MarkAsReadOnView
	case markReadOnView && markReadOnMediaPlayerCompletion:
		return MarkAsReadOnViewButWaitForPlayerCompletion
	case !markReadOnView && markReadOnMediaPlayerCompletion:
		return MarkAsReadOnlyOnPlayerCompletion
	case !markReadOnView && !markReadOnMediaPlayerCompletion:
		fallthrough // Explicit defaulting
	default:
		return NoAutoMarkAsRead
	}
}

// ExtractMarkAsReadBehavior returns the MarkReadOnView and MarkReadOnMediaPlayerCompletion values from the given MarkReadBehavior.
// Useful to extract the values from the form to the User model
func ExtractMarkAsReadBehavior(behavior MarkReadBehavior) (markReadOnView, markReadOnMediaPlayerCompletion bool) {
	switch behavior {
	case MarkAsReadOnView:
		return true, false
	case MarkAsReadOnViewButWaitForPlayerCompletion:
		return true, true
	case MarkAsReadOnlyOnPlayerCompletion:
		return false, true
	case NoAutoMarkAsRead:
		fallthrough // Explicit defaulting
	default:
		return false, false
	}
}

// UserExport holds user data for exporting to JSON.
type UserExport struct {
	UserModificationRequest
	Version string `json:"miniflux_version"`
}

// NewUserExport creates a new UserExport for exporting user data to JSON.
func NewUserExport(u *User) UserExport {
	return UserExport{
		UserModificationRequest: UserModificationRequest{
			Username:                        OptionalString(u.Username),
			Theme:                           OptionalString(u.Theme),
			Language:                        OptionalString(u.Language),
			Timezone:                        OptionalString(u.Timezone),
			EntryDirection:                  OptionalString(u.EntryDirection),
			EntryOrder:                      OptionalString(u.EntryOrder),
			Stylesheet:                      OptionalString(u.Stylesheet),
			CustomJS:                        OptionalString(u.CustomJS),
			ExternalFontHosts:               OptionalString(u.ExternalFontHosts),
			GoogleID:                        OptionalString(u.GoogleID),
			OpenIDConnectID:                 OptionalString(u.OpenIDConnectID),
			EntriesPerPage:                  OptionalNumber(u.EntriesPerPage),
			IsAdmin:                         OptionalField(u.IsAdmin),
			KeyboardShortcuts:               OptionalField(u.KeyboardShortcuts),
			ShowReadingTime:                 OptionalField(u.ShowReadingTime),
			EntrySwipe:                      OptionalField(u.EntrySwipe),
			GestureNav:                      OptionalString(u.GestureNav),
			DisplayMode:                     OptionalString(u.DisplayMode),
			DefaultReadingSpeed:             OptionalNumber(u.DefaultReadingSpeed),
			CJKReadingSpeed:                 OptionalNumber(u.CJKReadingSpeed),
			DefaultHomePage:                 OptionalString(u.DefaultHomePage),
			CategoriesSortingOrder:          OptionalString(u.CategoriesSortingOrder),
			MarkReadOnView:                  OptionalField(u.MarkReadOnView),
			MarkReadOnMediaPlayerCompletion: OptionalField(u.MarkReadOnMediaPlayerCompletion),
			MediaPlaybackRate:               OptionalField(u.MediaPlaybackRate),
			BlockFilterEntryRules:           OptionalString(u.BlockFilterEntryRules),
			KeepFilterEntryRules:            OptionalString(u.KeepFilterEntryRules),
			AlwaysOpenExternalLinks:         OptionalField(u.AlwaysOpenExternalLinks),
			OpenExternalLinksInNewTab:       OptionalField(u.OpenExternalLinksInNewTab),
		},
		Version: version.Version,
	}
}
