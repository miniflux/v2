// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import "miniflux.app/v2/internal/version"

// Opts holds parsed configuration options.
var Opts *configOptions

var defaultHTTPClientUserAgent = "Mozilla/5.0 (compatible; Miniflux/" + version.Version + "; +https://miniflux.app)"
