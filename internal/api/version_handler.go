// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"
	"runtime"

	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/version"
)

func (h *handler) versionHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, r, &versionResponse{
		Version:   version.Version,
		Commit:    version.Commit,
		BuildDate: version.BuildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	})
}
