// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import (
	"embed"
	"encoding/json"
	"fmt"
)

type translationDict map[string]interface{}
type catalog map[string]translationDict

var defaultCatalog catalog

//go:embed translations/*.json
var translationFiles embed.FS

// LoadCatalogMessages loads and parses all translations encoded in JSON.
func LoadCatalogMessages() error {
	var err error
	defaultCatalog = make(catalog)

	for language := range AvailableLanguages() {
		defaultCatalog[language], err = loadTranslationFile(language)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadTranslationFile(language string) (translationDict, error) {
	translationFileData, err := translationFiles.ReadFile(fmt.Sprintf("translations/%s.json", language))
	if err != nil {
		return nil, err
	}

	translationMessages, err := parseTranslationMessages(translationFileData)
	if err != nil {
		return nil, err
	}

	return translationMessages, nil
}

func parseTranslationMessages(data []byte) (translationDict, error) {
	var translationMessages translationDict
	if err := json.Unmarshal(data, &translationMessages); err != nil {
		return nil, fmt.Errorf(`invalid translation file: %w`, err)
	}
	return translationMessages, nil
}
