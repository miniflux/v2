// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package version // import "miniflux.app/v2/internal/version"

import (
	"runtime/debug"
	"time"
)

// Variables populated at build time.
var (
	Version   = "Development Version"
	Commit    = getCommit()
	BuildDate = getBuildDate()
)

func getCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:8] // Short commit hash
			}
		}
	}
	return "HEAD"
}

func getBuildDate() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return time.Now().Format(time.RFC3339)
}
