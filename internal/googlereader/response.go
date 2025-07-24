// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"fmt"
	"net/http"
)

type loginResponse struct {
	SID  string `json:"SID,omitempty"`
	LSID string `json:"LSID,omitempty"`
	Auth string `json:"Auth,omitempty"`
}

func (l loginResponse) String() string {
	return fmt.Sprintf("SID=%s\nLSID=%s\nAuth=%s\n", l.SID, l.LSID, l.Auth)
}

type userInfoResponse struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserProfileID string `json:"userProfileId"`
	UserEmail     string `json:"userEmail"`
}

type subscriptionResponse struct {
	ID         string                         `json:"id"`
	Title      string                         `json:"title"`
	Categories []subscriptionCategoryResponse `json:"categories"`
	URL        string                         `json:"url"`
	HTMLURL    string                         `json:"htmlUrl"`
	IconURL    string                         `json:"iconUrl"`
}

type subscriptionsResponse struct {
	Subscriptions []subscriptionResponse `json:"subscriptions"`
}

type quickAddResponse struct {
	NumResults int64  `json:"numResults"`
	Query      string `json:"query,omitempty"`
	StreamID   string `json:"streamId,omitempty"`
	StreamName string `json:"streamName,omitempty"`
}

type subscriptionCategoryResponse struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Type  string `json:"type,omitempty"`
}

type itemRef struct {
	ID              string `json:"id"`
	DirectStreamIDs string `json:"directStreamIds,omitempty"`
	TimestampUsec   string `json:"timestampUsec,omitempty"`
}

type streamIDResponse struct {
	ItemRefs     []itemRef `json:"itemRefs"`
	Continuation int       `json:"continuation,omitempty,string"`
}

type tagsResponse struct {
	Tags []subscriptionCategoryResponse `json:"tags"`
}

type streamContentItemsResponse struct {
	Direction string        `json:"direction"`
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Self      []contentHREF `json:"self"`
	Updated   int64         `json:"updated"`
	Items     []contentItem `json:"items"`
	Author    string        `json:"author"`
}

type contentItem struct {
	ID            string                 `json:"id"`
	Categories    []string               `json:"categories"`
	Title         string                 `json:"title"`
	CrawlTimeMsec string                 `json:"crawlTimeMsec"`
	TimestampUsec string                 `json:"timestampUsec"`
	Published     int64                  `json:"published"`
	Updated       int64                  `json:"updated"`
	Author        string                 `json:"author"`
	Alternate     []contentHREFType      `json:"alternate"`
	Summary       contentItemContent     `json:"summary"`
	Content       contentItemContent     `json:"content"`
	Origin        contentItemOrigin      `json:"origin"`
	Enclosure     []contentItemEnclosure `json:"enclosure"`
	Canonical     []contentHREF          `json:"canonical"`
}

type contentHREFType struct {
	HREF string `json:"href"`
	Type string `json:"type"`
}

type contentHREF struct {
	HREF string `json:"href"`
}

type contentItemEnclosure struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}
type contentItemContent struct {
	Direction string `json:"direction"`
	Content   string `json:"content"`
}

type contentItemOrigin struct {
	StreamID string `json:"streamId"`
	Title    string `json:"title"`
	HTMLUrl  string `json:"htmlUrl"`
}

func sendUnauthorizedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Reader-Google-Bad-Token", "true")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func sendOkayResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
