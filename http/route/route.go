// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package route // import "miniflux.app/http/route"

import (
	"strconv"

	"github.com/gorilla/mux"
	"miniflux.app/logger"
)

// Path returns the defined route based on given arguments.
func Path(router *mux.Router, name string, args ...interface{}) string {
	route := router.Get(name)
	if route == nil {
		logger.Fatal("[Route] Route not found: %s", name)
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
		logger.Fatal("[Route] %v", err)
	}

	return result.String()
}
