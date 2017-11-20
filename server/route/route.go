// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package route

import (
	"log"
	"strconv"

	"github.com/gorilla/mux"
)

func GetRoute(router *mux.Router, name string, args ...interface{}) string {
	route := router.Get(name)
	if route == nil {
		log.Fatalln("Route not found:", name)
	}

	var pairs []string
	for _, param := range args {
		switch param.(type) {
		case string:
			pairs = append(pairs, param.(string))
		case int64:
			val := param.(int64)
			pairs = append(pairs, strconv.FormatInt(val, 10))
		}
	}

	result, err := route.URLPath(pairs...)
	if err != nil {
		log.Fatalln(err)
	}

	return result.String()
}
