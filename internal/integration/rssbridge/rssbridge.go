// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssbridge // import "miniflux.app/integration/rssbridge"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Bridge struct {
	URL        string     `json:"url"`
	BridgeMeta BridgeMeta `json:"bridgeMeta"`
}

type BridgeMeta struct {
	Name string `json:"name"`
}

func DetectBridges(rssbridgeURL, websiteURL string) (bridgeResponse []Bridge, err error) {
	u, err := url.Parse(rssbridgeURL)
	if err != nil {
		return nil, err
	}
	values := u.Query()
	values.Add("action", "findfeed")
	values.Add("format", "atom")
	values.Add("url", websiteURL)
	u.RawQuery = values.Encode()

	response, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to excute request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return
	}
	if response.StatusCode > 400 {
		return nil, fmt.Errorf("RSS-Bridge: unexpected status code %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(&bridgeResponse); err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to decode bridge response: %w", err)
	}
	return
}
