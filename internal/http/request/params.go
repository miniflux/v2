// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// FormInt64Value returns the named form value parsed as int64, or 0 on error.
func FormInt64Value(r *http.Request, param string) int64 {
	value := r.FormValue(param)
	integer, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return integer
}

// RouteInt64Param returns the named route parameter parsed as int64, or 0 when missing or invalid.
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

// RouteStringParam returns the named route parameter as a string.
func RouteStringParam(r *http.Request, param string) string {
	vars := mux.Vars(r)
	return vars[param]
}

// QueryStringParam returns the named query parameter, or defaultValue if it is empty.
func QueryStringParam(r *http.Request, param, defaultValue string) string {
	value := r.URL.Query().Get(param)
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryStringParamList returns the non-empty, trimmed values for the named query parameter.
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

// QueryIntParam returns the named query parameter parsed as int, or defaultValue when missing, invalid, or negative.
func QueryIntParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return defaultValue
	}

	if val < 0 {
		return defaultValue
	}

	return int(val)
}

// QueryInt64Param returns the named query parameter parsed as int64, or defaultValue when missing, invalid, or negative.
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

// QueryBoolParam returns the named query parameter parsed as bool, or defaultValue when missing or invalid.
func QueryBoolParam(r *http.Request, param string, defaultValue bool) bool {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	val, err := strconv.ParseBool(value)

	if err != nil {
		return defaultValue
	}

	return val
}

// HasQueryParam reports whether the query string contains the named parameter.
func HasQueryParam(r *http.Request, param string) bool {
	values := r.URL.Query()
	_, ok := values[param]
	return ok
}
