// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package browser // import "miniflux.app/reader/browser"

import (
	"miniflux.app/errors"
	"miniflux.app/http/client"
)

var (
	errRequestFailed    = "Unable to open this link: %v"
	errServerFailure    = "Unable to fetch this resource (Status Code = %d)"
	errEncoding         = "Unable to normalize encoding: %q"
	errEmptyFeed        = "This feed is empty"
	errResourceNotFound = "Resource not found (404), this feed doesn't exist anymore, check the feed URL"
	errNotAuthorized    = "You are not authorized to access this resource (invalid username/password)"
)

// Exec executes a HTTP request and handles errors.
func Exec(request *client.Client) (*client.Response, *errors.LocalizedError) {
	response, err := request.Get()
	if err != nil {
		if e, ok := err.(*errors.LocalizedError); ok {
			return nil, e
		}
		return nil, errors.NewLocalizedError(errRequestFailed, err)
	}

	if response.IsNotFound() {
		return nil, errors.NewLocalizedError(errResourceNotFound)
	}

	if response.IsNotAuthorized() {
		return nil, errors.NewLocalizedError(errNotAuthorized)
	}

	if response.HasServerFailure() {
		return nil, errors.NewLocalizedError(errServerFailure, response.StatusCode)
	}

	if response.StatusCode != 304 {
		// Content-Length = -1 when no Content-Length header is sent.
		if response.ContentLength == 0 {
			return nil, errors.NewLocalizedError(errEmptyFeed)
		}

		if err := response.EnsureUnicodeBody(); err != nil {
			return nil, errors.NewLocalizedError(errEncoding, err)
		}
	}

	return response, nil
}
