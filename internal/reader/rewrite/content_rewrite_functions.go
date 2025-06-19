// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"encoding/base64"
	"fmt"
	"html"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"miniflux.app/v2/internal/config"

	nethtml "golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
)

var (
	youtubeVideoRegex = regexp.MustCompile(`youtube\.com/watch\?v=(.*)$`)
	youtubeShortRegex = regexp.MustCompile(`youtube\.com/shorts/([a-zA-Z0-9_-]{11})$`)
	youtubeIdRegex    = regexp.MustCompile(`youtube_id"?\s*[:=]\s*"([a-zA-Z0-9_-]{11})"`)
	invidioRegex      = regexp.MustCompile(`https?://(.*)/watch\?v=(.*)`)
	textLinkRegex     = regexp.MustCompile(`(?mi)(\bhttps?:\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])`)
)

// titlelize returns a copy of the string s with all Unicode letters that begin words
// mapped to their Unicode title case.
func titlelize(s string) string {
	// A closure is used here to remember the previous character
	// so that we can check if there is a space preceding the current
	// character.
	previous := ' '
	return strings.Map(
		func(current rune) rune {
			if unicode.IsSpace(previous) {
				previous = current
				return unicode.ToTitle(current)
			}
			previous = current
			return current
		}, strings.ToLower(s))
}

func addImageTitle(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	matches := doc.Find("img[src][title]")

	if matches.Length() > 0 {
		matches.Each(func(i int, img *goquery.Selection) {
			altAttr := img.AttrOr("alt", "")
			srcAttr, _ := img.Attr("src")
			titleAttr, _ := img.Attr("title")

			img.ReplaceWithHtml(`<figure><img src="` + srcAttr + `" alt="` + altAttr + `"/><figcaption><p>` + html.EscapeString(titleAttr) + `</p></figcaption></figure>`)
		})

		output, _ := doc.FindMatcher(goquery.Single("body")).Html()
		return output
	}

	return entryContent
}

func addMailtoSubject(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	matches := doc.Find(`a[href^="mailto:"]`)

	if matches.Length() > 0 {
		matches.Each(func(i int, a *goquery.Selection) {
			hrefAttr, _ := a.Attr("href")

			mailto, err := url.Parse(hrefAttr)
			if err != nil {
				return
			}

			subject := mailto.Query().Get("subject")
			if subject == "" {
				return
			}

			a.AppendHtml(" [" + html.EscapeString(subject) + "]")
		})

		output, _ := doc.FindMatcher(goquery.Single("body")).Html()
		return output
	}

	return entryContent
}

func addDynamicImage(entryContent string) string {
	parserHtml, err := nethtml.ParseWithOptions(strings.NewReader(entryContent), nethtml.ParseOptionEnableScripting(false))
	if err != nil {
		return entryContent
	}
	doc := goquery.NewDocumentFromNode(parserHtml)

	// Ordered most preferred to least preferred.
	candidateAttrs := []string{
		"data-src",
		"data-original",
		"data-orig",
		"data-url",
		"data-orig-file",
		"data-large-file",
		"data-medium-file",
		"data-original-mos",
		"data-2000src",
		"data-1000src",
		"data-800src",
		"data-655src",
		"data-500src",
		"data-380src",
	}

	candidateSrcsetAttrs := []string{
		"data-srcset",
	}

	changed := false

	doc.Find("img,div").Each(func(i int, img *goquery.Selection) {
		// Src-linked candidates
		for _, candidateAttr := range candidateAttrs {
			if srcAttr, found := img.Attr(candidateAttr); found {
				changed = true

				if img.Is("img") {
					img.SetAttr("src", srcAttr)
				} else {
					altAttr := img.AttrOr("alt", "")
					img.ReplaceWithHtml(`<img src="` + srcAttr + `" alt="` + altAttr + `"/>`)
				}

				break
			}
		}

		// Srcset-linked candidates
		for _, candidateAttr := range candidateSrcsetAttrs {
			if srcAttr, found := img.Attr(candidateAttr); found {
				changed = true

				if img.Is("img") {
					img.SetAttr("srcset", srcAttr)
				} else {
					altAttr := img.AttrOr("alt", "")
					img.ReplaceWithHtml(`<img srcset="` + srcAttr + `" alt="` + altAttr + `"/>`)
				}

				break
			}
		}
	})

	if !changed {
		doc.Find("noscript").Each(func(i int, noscript *goquery.Selection) {
			if img := noscript.Find("img"); img.Length() == 1 {
				img.Unwrap()
				changed = true
			}
		})
	}

	if changed {
		output, _ := doc.FindMatcher(goquery.Single("body")).Html()
		return output
	}

	return entryContent
}

func addDynamicIframe(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	// Ordered most preferred to least preferred.
	candidateAttrs := []string{
		"data-src",
		"data-original",
		"data-orig",
		"data-url",
		"data-lazy-src",
	}

	changed := false

	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		for _, candidateAttr := range candidateAttrs {
			if srcAttr, found := iframe.Attr(candidateAttr); found {
				changed = true

				iframe.SetAttr("src", srcAttr)

				break
			}
		}
	})

	if changed {
		output, _ := doc.FindMatcher(goquery.Single("body")).Html()
		return output
	}

	return entryContent
}

func fixMediumImages(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	doc.Find("figure.paragraph-image").Each(func(i int, paragraphImage *goquery.Selection) {
		noscriptElement := paragraphImage.Find("noscript")
		if noscriptElement.Length() > 0 {
			paragraphImage.ReplaceWithHtml(noscriptElement.Text())
		}
	})

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

func useNoScriptImages(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	doc.Find("figure").Each(func(i int, figureElement *goquery.Selection) {
		imgElement := figureElement.Find("img")
		if imgElement.Length() > 0 {
			noscriptElement := figureElement.Find("noscript")
			if noscriptElement.Length() > 0 {
				figureElement.PrependHtml(noscriptElement.Text())
				imgElement.Remove()
				noscriptElement.Remove()
			}
		}
	})

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

func getYoutubVideoIDFromURL(entryURL string) string {
	matches := youtubeVideoRegex.FindStringSubmatch(entryURL)

	if len(matches) != 2 {
		matches = youtubeShortRegex.FindStringSubmatch(entryURL)
	}

	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func addVideoPlayerIframe(absoluteVideoURL, entryContent string) string {
	video := `<iframe width="650" height="350" frameborder="0" src="` + absoluteVideoURL + `" allowfullscreen></iframe>`
	return video + `<br>` + entryContent
}

func addYoutubeVideoRewriteRule(entryURL, entryContent string) string {
	if videoURL := getYoutubVideoIDFromURL(entryURL); videoURL != "" {
		return addVideoPlayerIframe(config.Opts.YouTubeEmbedUrlOverride()+videoURL, entryContent)
	}
	return entryContent
}

func addYoutubeVideoUsingInvidiousPlayer(entryURL, entryContent string) string {
	if videoURL := getYoutubVideoIDFromURL(entryURL); videoURL != "" {
		return addVideoPlayerIframe(`https://`+config.Opts.InvidiousInstance()+`/embed/`+videoURL, entryContent)
	}
	return entryContent
}

func addYoutubeVideoFromId(entryContent string) string {
	matches := youtubeIdRegex.FindAllStringSubmatch(entryContent, -1)
	if matches == nil {
		return entryContent
	}
	sb := strings.Builder{}
	for _, match := range matches {
		if len(match) == 2 {
			sb.WriteString(`<iframe width="650" height="350" frameborder="0" src="`)
			sb.WriteString(config.Opts.YouTubeEmbedUrlOverride())
			sb.WriteString(match[1])
			sb.WriteString(`" allowfullscreen></iframe><br>`)
		}
	}
	sb.WriteString(entryContent)
	return sb.String()
}

func addInvidiousVideo(entryURL, entryContent string) string {
	matches := invidioRegex.FindStringSubmatch(entryURL)
	if len(matches) == 3 {
		video := `<iframe width="650" height="350" frameborder="0" src="https://` + matches[1] + `/embed/` + matches[2] + `" allowfullscreen></iframe>`
		return video + `<br>` + entryContent
	}
	return entryContent
}

func addPDFLink(entryURL, entryContent string) string {
	if strings.HasSuffix(entryURL, ".pdf") {
		return fmt.Sprintf(`<a href=%q>PDF</a><br>%s`, entryURL, entryContent)
	}
	return entryContent
}

func replaceTextLinks(input string) string {
	return textLinkRegex.ReplaceAllString(input, `<a href="${1}">${1}</a>`)
}

func replaceCustom(entryContent string, searchTerm string, replaceTerm string) string {
	re, err := regexp.Compile(searchTerm)
	if err == nil {
		return re.ReplaceAllString(entryContent, replaceTerm)
	}
	return entryContent
}

func removeCustom(entryContent string, selector string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	doc.Find(selector).Remove()

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

func addCastopodEpisode(entryURL, entryContent string) string {
	player := `<iframe width="650" frameborder="0" src="` + entryURL + `/embed/light"></iframe>`

	return player + `<br>` + entryContent
}

func applyFuncOnTextContent(entryContent string, selector string, repl func(string) string) string {
	var treatChildren func(i int, s *goquery.Selection)
	treatChildren = func(i int, s *goquery.Selection) {
		if s.Nodes[0].Type == nethtml.TextNode {
			s.ReplaceWithHtml(repl(s.Nodes[0].Data))
		} else {
			s.Contents().Each(treatChildren)
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	doc.Find(selector).Each(treatChildren)

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

func decodeBase64Content(entryContent string) string {
	if ret, err := base64.StdEncoding.DecodeString(strings.TrimSpace(entryContent)); err != nil {
		return entryContent
	} else {
		return html.EscapeString(string(ret))
	}
}

func addHackerNewsLinksUsing(entryContent, app string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	hn_prefix := "https://news.ycombinator.com/"
	matches := doc.Find(`a[href^="` + hn_prefix + `"]`)

	if matches.Length() > 0 {
		matches.Each(func(i int, a *goquery.Selection) {
			hrefAttr, _ := a.Attr("href")

			hn_uri, err := url.Parse(hrefAttr)
			if err != nil {
				return
			}

			switch app {
			case "opener":
				params := url.Values{}
				params.Add("url", hn_uri.String())

				url := url.URL{
					Scheme:   "opener",
					Host:     "x-callback-url",
					Path:     "show-options",
					RawQuery: params.Encode(),
				}

				open_with_opener := `<a href="` + url.String() + `">Open with Opener</a>`
				a.Parent().AppendHtml(" " + open_with_opener)
			case "hack":
				url := strings.Replace(hn_uri.String(), hn_prefix, "hack://", 1)

				open_with_hack := `<a href="` + url + `">Open with HACK</a>`
				a.Parent().AppendHtml(" " + open_with_hack)
			default:
				slog.Warn("Unknown app provided for openHackerNewsLinksWith rewrite rule",
					slog.String("app", app),
				)
				return
			}
		})

		output, _ := doc.FindMatcher(goquery.Single("body")).Html()
		return output
	}

	return entryContent
}

func removeTables(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	selectors := []string{"table", "tbody", "thead", "td", "th", "td"}

	var loopElement *goquery.Selection

	for _, selector := range selectors {
		for {
			loopElement = doc.FindMatcher(goquery.Single(selector))

			if loopElement.Length() == 0 {
				break
			}

			innerHtml, err := loopElement.Html()
			if err != nil {
				break
			}

			loopElement.Parent().AppendHtml(innerHtml)
			loopElement.Remove()
		}
	}

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

func fixGhostCards(entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	const cardSelector = "figure.kg-card"
	var currentList *goquery.Selection

	doc.Find(cardSelector).Each(func(i int, s *goquery.Selection) {
		title := s.Find(".kg-bookmark-title").First().Text()
		author := s.Find(".kg-bookmark-author").First().Text()
		href := s.Find("a.kg-bookmark-container").First().AttrOr("href", "")

		// if there is no link or title, skip processing
		if href == "" || title == "" {
			return
		}

		link := ""
		if author == "" || strings.HasSuffix(title, author) {
			link = fmt.Sprintf("<a href=\"%s\">%s</a>", href, title)
		} else {
			link = fmt.Sprintf("<a href=\"%s\">%s - %s</a>", href, title, author)
		}

		next := s.Next()

		// if the next element is also a card, start a list
		if next.Is(cardSelector) && currentList == nil {
			currentList = s.BeforeHtml("<ul></ul>").Prev()
		}

		if currentList != nil {
			// add this card to the list, then delete it
			currentList.AppendHtml("<li>" + link + "</li>")
			s.Remove()
		} else {
			// replace single card
			s.ReplaceWithHtml(link)
		}

		// if the next element is not a card, start a new list
		if !next.Is(cardSelector) && currentList != nil {
			currentList = nil
		}
	})

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return strings.TrimSpace(output)
}
