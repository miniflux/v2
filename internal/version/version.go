// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package version // import "miniflux.app/v2/internal/version"

import (
	"runtime/debug"
)

// Variables populated at build time when using LD_FLAGS.
var (
	Version   = ""
	Commit    = ""
	BuildDate = ""
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

// Populate build information from VCS metadata if LDFLAGS are not set.
// Falls back to values from the Go module's build info when available.
func init() {
	if Version == "" {
		// Some Miniflux clients expect a specific version format.
		// For example, Flux News converts the string version to an integer.
		Version = "2.2.x-dev"
	}
	if Commit == "" {
		Commit = getCommit()
	}
	if BuildDate == "" {
		BuildDate = getBuildDate()
	}
}
