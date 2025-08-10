// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "influxeed-engine/v2/internal/reader/processor"

import (
	"influxeed-engine/v2/internal/config"
	"influxeed-engine/v2/internal/model"
	"influxeed-engine/v2/internal/urllib"
)

func shouldFetchNebulaWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchNebulaWatchTime() {
		return false
	}

	return urllib.DomainWithoutWWW(entry.URL) == "nebula.tv"
}

func fetchNebulaWatchTime(websiteURL string) (int, error) {
	return fetchWatchTime(websiteURL, `meta[property="video:duration"]`, false)
}
