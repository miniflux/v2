// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
)

// XMLResponse handles XML responses.
type XMLResponse struct {
	writer  http.ResponseWriter
	request *http.Request
}

// Download force the download of a XML document.
func (x *XMLResponse) Download(filename, data string) {
	x.writer.Header().Set("Content-Type", "text/xml")
	x.writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	x.writer.Write([]byte(data))
}

// Serve forces the XML to be sent to browser.
func (x *XMLResponse) Serve(data string) {
	x.writer.Header().Set("Content-Type", "text/xml")
	x.writer.Write([]byte(data))
}
