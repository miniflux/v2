// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package request

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Cookie returns the cookie value.
func Cookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err == http.ErrNoCookie {
		return ""
	}

	return cookie.Value
}

// FormIntValue returns a form value as integer.
func FormIntValue(r *http.Request, param string) int64 {
	value := r.FormValue(param)
	integer, _ := strconv.Atoi(value)
	return int64(integer)
}

// IntParam returns an URL route parameter as integer.
func IntParam(r *http.Request, param string) (int64, error) {
	vars := mux.Vars(r)
	value, err := strconv.Atoi(vars[param])
	if err != nil {
		return 0, fmt.Errorf("request: %s parameter is not an integer", param)
	}

	if value < 0 {
		return 0, nil
	}

	return int64(value), nil
}

// Param returns an URL route parameter as string.
func Param(r *http.Request, param, defaultValue string) string {
	vars := mux.Vars(r)
	value := vars[param]
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryParam returns a querystring parameter as string.
func QueryParam(r *http.Request, param, defaultValue string) string {
	value := r.URL.Query().Get(param)
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryIntParam returns a querystring parameter as integer.
func QueryIntParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if val < 0 {
		return defaultValue
	}

	return val
}

// HasQueryParam checks if the query string contains the given parameter.
func HasQueryParam(r *http.Request, param string) bool {
	values := r.URL.Query()
	_, ok := values[param]
	return ok
}

// RealIP returns client's real IP address.
func RealIP(r *http.Request) string {
	headers := []string{"X-Forwarded-For", "X-Real-Ip"}
	for _, header := range headers {
		value := r.Header.Get(header)

		if value != "" {
			addresses := strings.Split(value, ",")
			address := strings.TrimSpace(addresses[0])

			if net.ParseIP(address) != nil {
				return address
			}
		}
	}

	// Fallback to TCP/IP source IP address.
	var remoteIP string
	if strings.ContainsRune(r.RemoteAddr, ':') {
		remoteIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	} else {
		remoteIP = r.RemoteAddr
	}

	return remoteIP
}
