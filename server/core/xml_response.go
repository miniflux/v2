// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"net/http"
)

type XmlResponse struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (x *XmlResponse) Download(filename, data string) {
	x.writer.Header().Set("Content-Type", "text/xml")
	x.writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	x.writer.Write([]byte(data))
}
