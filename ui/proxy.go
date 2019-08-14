// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
)

func (h *handler) imageProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the image is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("No URL provided"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this URL"))
		return
	}

	req, err := http.NewRequest("GET", string(decodedURL), nil)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	req.Header.Add("User-Agent", client.DefaultUserAgent)
	req.Header.Add("Connection", "close")

	clt := &http.Client{
		Timeout: time.Duration(config.Opts.HTTPClientTimeout()) * time.Second,
	}

	resp, err := clt.Do(req)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		html.NotFound(w, r)
		return
	}

	etag := crypto.HashFromBytes(decodedURL)

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}
