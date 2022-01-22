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
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/logger"
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

	imageURL := string(decodedURL)
	logger.Debug(`[Proxy] Fetching %q`, imageURL)

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Note: User-Agent HTTP header is omitted to avoid being blocked by bot protection mechanisms.
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
		logger.Error(`[Proxy] Status Code is %d for URL %q`, resp.StatusCode, imageURL)
		html.NotFound(w, r)
		return
	}

	etag := crypto.HashFromBytes(decodedURL)

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Security-Policy", `default-src 'self'`)
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}
