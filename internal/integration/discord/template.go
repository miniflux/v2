// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package discord // import "miniflux.app/v2/internal/integration/discord"

import (
	"html"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"
)

type templateData struct {
	Title       string
	URL         string
	Description string
	Author      string
	Date        string
	FeedTitle   string
	FeedURL     string
}

func buildTemplateData(feed *model.Feed, entry *model.Entry) templateData {
	content := stripHTMLTags(entry.Content)
	content = html.UnescapeString(content)
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) > 500 {
		content = string(runes[:500]) + "..."
	}

	var dateStr string
	if !entry.Date.IsZero() {
		dateStr = entry.Date.Format(time.RFC3339)
	}

	return templateData{
		Title:       entry.Title,
		URL:         entry.URL,
		Description: content,
		Author:      entry.Author,
		Date:        dateStr,
		FeedTitle:   feed.Title,
		FeedURL:     feed.SiteURL,
	}
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func renderTemplate(tplStr string, data templateData) string {
	replacements := map[string]string{
		"{title}":       data.Title,
		"{url}":         data.URL,
		"{description}": data.Description,
		"{author}":      data.Author,
		"{date}":        data.Date,
		"{feed_title}":  data.FeedTitle,
		"{feed_url}":    data.FeedURL,
	}

	result := tplStr
	for token, value := range replacements {
		result = strings.ReplaceAll(result, token, value)
	}

	runes := []rune(result)
	if len(runes) > 4000 {
		result = string(runes[:4000])
	}

	return result
}

func ValidateTemplate(tplStr string) bool {
	recognized := map[string]bool{
		"{title}":       true,
		"{url}":         true,
		"{description}": true,
		"{author}":      true,
		"{date}":        true,
		"{feed_title}":  true,
		"{feed_url}":    true,
	}

	i := 0
	runes := []rune(tplStr)
	for i < len(runes) {
		if runes[i] == '{' {
			end := indexOf(runes, '}', i)
			if end == -1 {
				break
			}
			token := string(runes[i : end+1])
			if !recognized[token] {
				return false
			}
			i = end + 1
		} else {
			i++
		}
	}
	return true
}

func indexOf(runes []rune, target rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
}
