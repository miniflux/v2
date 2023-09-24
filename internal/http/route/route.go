// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package route // import "miniflux.app/v2/internal/http/route"

import (
	"strconv"

	"github.com/gorilla/mux"
)

// Path returns the defined route based on given arguments.
func Path(router *mux.Router, name string, args ...any) string {
	route := router.Get(name)
	if route == nil {
		panic("route not found: " + name)
	}

	var pairs []string
	for _, arg := range args {
		switch param := arg.(type) {
		case string:
			pairs = append(pairs, param)
		case int64:
			pairs = append(pairs, strconv.FormatInt(param, 10))
		}
	}

	result, err := route.URLPath(pairs...)
	if err != nil {
		panic(err)
	}

	return result.String()
}
