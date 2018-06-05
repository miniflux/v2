// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/miniflux/miniflux/crypto"
	"github.com/miniflux/miniflux/http/client"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/logger"
)

// ImageProxy fetch an image from a remote server and sent it back to the browser.
func (c *Controller) ImageProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header we assume the image in stored in browser cache
	if r.Header.Get("If-None-Match") != "" {
		response.NotModified(w)
		return
	}

	encodedURL := request.Param(r, "encodedURL", "")
	if encodedURL == "" {
		html.BadRequest(w, errors.New("No URL provided"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, errors.New("Unable to decode this URL"))
		return
	}

	clt := client.New(string(decodedURL))
	resp, err := clt.Get()
	if err != nil {
		logger.Error("[Controller:ImageProxy] %v", err)
		html.NotFound(w)
		return
	}

	if resp.HasServerFailure() {
		html.NotFound(w)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	etag := crypto.HashFromBytes(body)

	response.Cache(w, r, resp.ContentType, etag, body, 72*time.Hour)
}
