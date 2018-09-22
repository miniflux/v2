// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import (
	"encoding/json"
	"fmt"
)

type catalogMessages map[string]interface{}
type catalog map[string]catalogMessages

func parseCatalogMessages(data string) (catalogMessages, error) {
	var translations catalogMessages
	if err := json.Unmarshal([]byte(data), &translations); err != nil {
		return nil, fmt.Errorf("invalid translation file: %v", err)
	}
	return translations, nil
}
