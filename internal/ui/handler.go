// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"

	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/template"
	"miniflux.app/v2/internal/ui/static"
	"miniflux.app/v2/internal/worker"
)

type handler struct {
	basePath string
	store    *storage.Storage
	tpl      *template.Engine
	pool     *worker.Pool
}

func (h *handler) routePath(format string, args ...any) string {
	if len(args) > 0 {
		return h.basePath + fmt.Sprintf(format, args...)
	}
	return h.basePath + format
}

func (h *handler) iconPath(filename string) string {
	return h.basePath + fmt.Sprintf("/icon/%s/%s", static.BinaryBundles[filename].Checksum, filename)
}
