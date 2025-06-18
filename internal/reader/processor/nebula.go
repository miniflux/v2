// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
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
