// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import (
	"bytes"
	"sync"
	"testing"
)

// TestRenderConcurrency renders the same template concurrently in different
// languages. Because Render binds per-request, language-specific functions
// ("t", "plural", "elapsed") onto the template, doing so on a shared template
// while other requests execute it corrupts the output: a request can be served
// another request's language. Each concurrent render must match the output of
// the equivalent sequential render for its language.
func TestRenderConcurrency(t *testing.T) {
	engine := NewEngine("")
	engine.ParseTemplates()

	languages := []string{"en_US", "fr_FR", "de_DE", "es_ES", "pt_BR", "ru_RU", "zh_CN", "it_IT"}

	newData := func(language string) map[string]any {
		return map[string]any{"language": language, "theme": "system_serif"}
	}

	// Establish the expected output for each language sequentially.
	expected := make(map[string][]byte, len(languages))
	for _, language := range languages {
		expected[language] = engine.Render("offline.html", newData(language))
	}

	const iterations = 300

	var wg sync.WaitGroup
	var mu sync.Mutex
	mismatches := make(map[string]int)

	for i := 0; i < iterations; i++ {
		language := languages[i%len(languages)]
		wg.Add(1)
		go func(language string) {
			defer wg.Done()
			got := engine.Render("offline.html", newData(language))
			if !bytes.Equal(got, expected[language]) {
				mu.Lock()
				mismatches[language]++
				mu.Unlock()
			}
		}(language)
	}
	wg.Wait()

	if len(mismatches) > 0 {
		total := 0
		for _, n := range mismatches {
			total += n
		}
		t.Fatalf("concurrent Render produced wrong output for %d/%d requests (wrong-language translations); per-language mismatches: %v", total, iterations, mismatches)
	}
}
