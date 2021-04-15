// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package googlereader // import "miniflux.app/googlereader"

import "fmt"

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

type subscriptionCategory struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type subscriptionsResponse struct {
	Subscriptions []subscription `json:"subscriptions"`
}
