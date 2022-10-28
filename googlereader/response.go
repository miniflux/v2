// Copyright 2022 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package googlereader // import "miniflux.app/googlereader"

import (
	"fmt"
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/logger"
)

type login struct {
	SID  string `json:"SID,omitempty"`
	LSID string `json:"LSID,omitempty"`
	Auth string `json:"Auth,omitempty"`
}

func (l login) String() string {
	return fmt.Sprintf("SID=%s\nLSID=%s\nAuth=%s\n", l.SID, l.LSID, l.Auth)
}

type userInfo struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserProfileID string `json:"userProfileId"`
	UserEmail     string `json:"userEmail"`
}

type subscription struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Categories []subscriptionCategory `json:"categories"`
	URL        string                 `json:"url"`
	HTMLURL    string                 `json:"htmlUrl"`
	IconURL    string                 `json:"iconUrl"`
}

type quickAddResponse struct {
	NumResults int64  `json:"numResults"`
	Query      string `json:"query,omitempty"`
	StreamID   string `json:"streamId,omitempty"`
	StreamName string `json:"streamName,omitempty"`
}

type subscriptionCategory struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Type  string `json:"type,omitempty"`
}
type subscriptionsResponse struct {
	Subscriptions []subscription `json:"subscriptions"`
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
	Tags []subscriptionCategory `json:"tags"`
}

type streamContentItems struct {
	Direction string            `json:"direction"`
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Self      []contentHREF     `json:"self"`
	Alternate []contentHREFType `json:"alternate"`
	Updated   int64             `json:"updated"`
	Items     []contentItem     `json:"items"`
	Author    string            `json:"author"`
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

// Unauthorized sends a not authorized error to the client.
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	logger.Error("[HTTP:Unauthorized] %s", r.URL)

	builder := response.New(w, r)
	builder.WithStatus(http.StatusUnauthorized)
	builder.WithHeader("Content-Type", "text/plain")
	builder.WithHeader("X-Reader-Google-Bad-Token", "true")
	builder.WithBody("Unauthorized")
	builder.Write()
}

// OK sends a ok response to the client.
func OK(w http.ResponseWriter, r *http.Request) {
	logger.Info("[HTTP:OK] %s", r.URL)

	builder := response.New(w, r)
	builder.WithStatus(http.StatusOK)
	builder.WithHeader("Content-Type", "text/plain")
	builder.WithBody("OK")
	builder.Write()
}
