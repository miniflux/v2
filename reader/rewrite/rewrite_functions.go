// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	youtubeRegex  = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
	invidioRegex  = regexp.MustCompile(`https?:\/\/(.*)\/watch\?v=(.*)`)
	imgRegex      = regexp.MustCompile(`<img [^>]+>`)
	textLinkRegex = regexp.MustCompile(`(?mi)(\bhttps?:\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])`)
)

func addImageTitle(entryURL, entryContent string) string {
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

		output, _ := doc.Find("body").First().Html()
		return output
	}

	return entryContent
}

func addMailtoSubject(entryURL, entryContent string) string {
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

		output, _ := doc.Find("body").First().Html()
		return output
	}

	return entryContent
}

func addDynamicImage(entryURL, entryContent string) string {
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
		"data-orig-file",
		"data-large-file",
		"data-medium-file",
		"data-2000src",
		"data-1000src",
		"data-800src",
		"data-655src",
		"data-500src",
		"data-380src",
	}

	changed := false

	doc.Find("img,div").Each(func(i int, img *goquery.Selection) {
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
	})

	if !changed {
		doc.Find("noscript").Each(func(i int, noscript *goquery.Selection) {
			matches := imgRegex.FindAllString(noscript.Text(), 2)

			if len(matches) == 1 {
				changed = true

				noscript.ReplaceWithHtml(matches[0])
			}
		})
	}

	if changed {
		output, _ := doc.Find("body").First().Html()
		return output
	}

	return entryContent
}

func fixMediumImages(entryURL, entryContent string) string {
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

	output, _ := doc.Find("body").First().Html()
	return output
}

func useNoScriptImages(entryURL, entryContent string) string {
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

	output, _ := doc.Find("body").First().Html()
	return output
}

func addYoutubeVideo(entryURL, entryContent string) string {
	matches := youtubeRegex.FindStringSubmatch(entryURL)

	if len(matches) == 2 {
		video := `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/` + matches[1] + `" allowfullscreen></iframe>`
		return video + `<br>` + entryContent
	}
	return entryContent
}

func addYoutubeVideoUsingInvidiousPlayer(entryURL, entryContent string) string {
	matches := youtubeRegex.FindStringSubmatch(entryURL)

	if len(matches) == 2 {
		video := `<iframe width="650" height="350" frameborder="0" src="https://invidio.us/embed/` + matches[1] + `" allowfullscreen></iframe>`
		return video + `<br>` + entryContent
	}
	return entryContent
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
		return fmt.Sprintf(`<a href="%s">PDF</a><br>%s`, entryURL, entryContent)
	}
	return entryContent
}

func replaceTextLinks(input string) string {
	return textLinkRegex.ReplaceAllString(input, `<a href="${1}">${1}</a>`)
}

func replaceLineFeeds(input string) string {
	return strings.Replace(input, "\n", "<br>", -1)
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

	output, _ := doc.Find("body").First().Html()
	return output
}
