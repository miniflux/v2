// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"
	"time"

	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/ui/static"
)

// Javascript renders application client side code.
func (c *Controller) Javascript(w http.ResponseWriter, r *http.Request) {
	response.Cache(w, r, "text/javascript; charset=utf-8", static.JavascriptChecksums["app"], []byte(static.Javascript["app"]), 48*time.Hour)
}
