// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package xml // import "miniflux.app/http/response/xml"

import (
	"fmt"
	"net/http"
)

// OK sends a XML document.
func OK(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(data))
}

// Attachment forces the download of a XML document.
func Attachment(w http.ResponseWriter, filename, data string) {
	w.Header().Set("Content-Type", "text/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte(data))
}
