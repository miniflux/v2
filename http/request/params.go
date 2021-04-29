// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package request // import "miniflux.app/http/request"

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// FormInt64Value returns a form value as integer.
func FormInt64Value(r *http.Request, param string) int64 {
	value := r.FormValue(param)
	integer, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return integer
}

// RouteInt64Param returns an URL route parameter as int64.
func RouteInt64Param(r *http.Request, param string) int64 {
	vars := mux.Vars(r)
	value, err := strconv.ParseInt(vars[param], 10, 64)
	if err != nil {
		return 0
	}

	if value < 0 {
		return 0
	}

	return value
}

// RouteStringParam returns a URL route parameter as string.
func RouteStringParam(r *http.Request, param string) string {
	vars := mux.Vars(r)
	return vars[param]
}

// QueryStringParam returns a query string parameter as string.
func QueryStringParam(r *http.Request, param, defaultValue string) string {
	value := r.URL.Query().Get(param)
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryStringParamList returns all values associated to the parameter.
func QueryStringParamList(r *http.Request, param string) []string {
	var results []string
	values := r.URL.Query()

	if _, found := values[param]; found {
		for _, value := range values[param] {
			value = strings.TrimSpace(value)
			if value != "" {
				results = append(results, value)
			}
		}
	}

	return results
}

// QueryIntParam returns a query string parameter as integer.
func QueryIntParam(r *http.Request, param string, defaultValue int) int {
	return int(QueryInt64Param(r, param, int64(defaultValue)))
}

// QueryInt64Param returns a query string parameter as int64.
func QueryInt64Param(r *http.Request, param string, defaultValue int64) int64 {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	if val < 0 {
		return defaultValue
	}

	return val
}

// QueryBooleanParam returns a query string parameter as boolean.
func QueryBooleanParam(r *http.Request, param string) bool {
	value := r.URL.Query().Get(param)
	if value == "" {
		return false
	}

	val, err := strconv.ParseBool(value)
	if err != nil {
		return true
	}

	return val
}

// QueryTimestampParam returns a query string (unix seconds) parameter as timestamp.
func QueryTimestampParam(r *http.Request, param string) *time.Time {
	value, err := strconv.ParseInt(r.URL.Query().Get(param), 10, 64)
	if err != nil {
		return nil
	}

	t := time.Unix(value, 0)

	return &t
}

// HasQueryParam checks if the query string contains the given parameter.
func HasQueryParam(r *http.Request, param string) bool {
	values := r.URL.Query()
	_, ok := values[param]
	return ok
}
