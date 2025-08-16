// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package version // import "miniflux.app/v2/internal/version"

import (
	"runtime/debug"
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
				if len(setting.Value) >= 8 {
					return setting.Value[:8]
				}
				return setting.Value
			}
		}
	}
	return "Unknown (built outside VCS)"
}

func getBuildDate() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return "Unknown (built outside VCS)"
}
