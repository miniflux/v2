// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import (
	"embed"
	"encoding/json"
	"fmt"
)

type translationDict struct {
	singulars map[string]string
	plurals   map[string][]string
}
type catalog map[string]translationDict

var defaultCatalog = make(catalog, len(AvailableLanguages))

//go:embed translations/*.json
var translationFiles embed.FS

func getTranslationDict(language string) (translationDict, error) {
	if _, ok := defaultCatalog[language]; !ok {
		var err error
		if defaultCatalog[language], err = loadTranslationFile(language); err != nil {
			return translationDict{}, err
		}
	}
	return defaultCatalog[language], nil
}

func loadTranslationFile(language string) (translationDict, error) {
	translationFileData, err := translationFiles.ReadFile("translations/" + language + ".json")
	if err != nil {
		return translationDict{}, err
	}

	translationMessages, err := parseTranslationMessages(translationFileData)
	if err != nil {
		return translationDict{}, err
	}

	return translationMessages, nil
}

func (t *translationDict) UnmarshalJSON(data []byte) error {
	var tmpMap map[string]any
	err := json.Unmarshal(data, &tmpMap)
	if err != nil {
		return err
	}

	m := translationDict{
		singulars: make(map[string]string),
		plurals:   make(map[string][]string),
	}

	for key, value := range tmpMap {
		switch vtype := value.(type) {
		case string:
			m.singulars[key] = vtype
		case []any:
			for _, translation := range vtype {
				if translationStr, ok := translation.(string); ok {
					m.plurals[key] = append(m.plurals[key], translationStr)
				} else {
					return fmt.Errorf("invalid type for translation in an array: %v", translation)
				}
			}
		default:
			return fmt.Errorf("invalid type (%T) for translation: %v", vtype, value)
		}
	}

	*t = m

	return nil
}

func parseTranslationMessages(data []byte) (translationDict, error) {
	var translationMessages translationDict
	if err := json.Unmarshal(data, &translationMessages); err != nil {
		return translationDict{}, fmt.Errorf(`invalid translation file: %w`, err)
	}
	return translationMessages, nil
}
