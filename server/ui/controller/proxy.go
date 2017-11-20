// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"encoding/base64"
	"errors"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/server/core"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func (c *Controller) ImageProxy(ctx *core.Context, request *core.Request, response *core.Response) {
	encodedURL := request.GetStringParam("encodedURL", "")
	if encodedURL == "" {
		response.Html().BadRequest(errors.New("No URL provided"))
		return
	}

	decodedURL, err := base64.StdEncoding.DecodeString(encodedURL)
	if err != nil {
		response.Html().BadRequest(errors.New("Unable to decode this URL"))
		return
	}

	resp, err := http.Get(string(decodedURL))
	if err != nil {
		log.Println(err)
		response.Html().NotFound()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.Html().NotFound()
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	etag := helper.HashFromBytes(body)
	contentType := resp.Header.Get("Content-Type")

	response.Cache(contentType, etag, body, 72*time.Hour)
}
