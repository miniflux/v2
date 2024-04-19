// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssbridge // import "miniflux.app/v2/internal/integration/rssbridge"

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultClientTimeout = 30 * time.Second

type Bridge struct {
	URL        string     `json:"url"`
	BridgeMeta BridgeMeta `json:"bridgeMeta"`
}

type BridgeMeta struct {
	Name string `json:"name"`
}

func DetectBridges(rssBridgeURL, websiteURL string) ([]*Bridge, error) {
	endpointURL, err := url.Parse(rssBridgeURL)
	if err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to parse bridge URL: %w", err)
	}

	values := endpointURL.Query()
	values.Add("action", "findfeed")
	values.Add("format", "atom")
	values.Add("url", websiteURL)
	endpointURL.RawQuery = values.Encode()

	slog.Debug("Detecting RSS bridges", slog.String("url", endpointURL.String()))

	request, err := http.NewRequest(http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to create request: %w", err)
	}

	httpClient := &http.Client{Timeout: defaultClientTimeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if response.StatusCode > 400 {
		return nil, fmt.Errorf("RSS-Bridge: unexpected status code %d", response.StatusCode)
	}

	var bridgeResponse []*Bridge
	if err := json.NewDecoder(response.Body).Decode(&bridgeResponse); err != nil {
		return nil, fmt.Errorf("RSS-Bridge: unable to decode bridge response: %w", err)
	}

	for _, bridge := range bridgeResponse {
		slog.Debug("Found RSS bridge",
			slog.String("name", bridge.BridgeMeta.Name),
			slog.String("url", bridge.URL),
		)

		if strings.HasPrefix(bridge.URL, "./") {
			bridge.URL = rssBridgeURL + bridge.URL[2:]

			slog.Debug("Rewrited relative RSS bridge URL",
				slog.String("name", bridge.BridgeMeta.Name),
				slog.String("url", bridge.URL),
			)
		}
	}

	return bridgeResponse, nil
}
