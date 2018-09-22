// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import (
	"encoding/json"
	"fmt"
)

type translationDict map[string]interface{}
type catalog map[string]translationDict

var defaultCatalog catalog

func init() {
	defaultCatalog = make(catalog)

	for language, data := range translations {
		messages, err := parseTranslationDict(data)
		if err != nil {
			panic(err)
		}

		defaultCatalog[language] = messages
	}
}

func parseTranslationDict(data string) (translationDict, error) {
	var translations translationDict
	if err := json.Unmarshal([]byte(data), &translations); err != nil {
		return nil, fmt.Errorf("invalid translation file: %v", err)
	}
	return translations, nil
}
