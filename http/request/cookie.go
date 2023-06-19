// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/http/request"

import "net/http"

// CookieValue returns the cookie value.
func CookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err == http.ErrNoCookie {
		return ""
	}

	return cookie.Value
}
