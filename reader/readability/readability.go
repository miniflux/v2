// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/reader/readability"

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"

	"miniflux.app/logger"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const (
	defaultTagsToScore = "section,h2,h3,h4,h5,h6,p,td,pre,div"
)

var (
	divToPElementsRegexp = regexp.MustCompile(`(?i)<(a|blockquote|dl|div|img|ol|p|pre|table|ul)`)
	sentenceRegexp       = regexp.MustCompile(`\.( |$)`)

	blacklistCandidatesRegexp  = regexp.MustCompile(`(?i)popupbody|-ad|g-plus`)
	okMaybeItsACandidateRegexp = regexp.MustCompile(`(?i)and|article|body|column|main|shadow`)
	unlikelyCandidatesRegexp   = regexp.MustCompile(`(?i)banner|breadcrumbs|combx|comment|community|cover-wrap|disqus|extra|foot|header|legends|menu|modal|related|remark|replies|rss|shoutbox|sidebar|skyscraper|social|sponsor|supplemental|ad-break|agegate|pagination|pager|popup|yom-remote`)

	negativeRegexp = regexp.MustCompile(`(?i)hidden|^hid$|hid$|hid|^hid |banner|combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|modal|outbrain|promo|related|scroll|share|shoutbox|sidebar|skyscraper|sponsor|shopping|tags|tool|widget|byline|author|dateline|writtenby|p-author`)
	positiveRegexp = regexp.MustCompile(`(?i)article|body|content|entry|hentry|h-entry|main|page|pagination|post|text|blog|story`)
)

type candidate struct {
	selection *goquery.Selection
	score     float32
}

func (c *candidate) Node() *html.Node {
	return c.selection.Get(0)
}

func (c *candidate) String() string {
	id, _ := c.selection.Attr("id")
	class, _ := c.selection.Attr("class")

	if id != "" && class != "" {
		return fmt.Sprintf("%s#%s.%s => %f", c.Node().DataAtom, id, class, c.score)
	} else if id != "" {
		return fmt.Sprintf("%s#%s => %f", c.Node().DataAtom, id, c.score)
	} else if class != "" {
		return fmt.Sprintf("%s.%s => %f", c.Node().DataAtom, class, c.score)
	}

	return fmt.Sprintf("%s => %f", c.Node().DataAtom, c.score)
}

type candidateList map[*html.Node]*candidate

func (c candidateList) String() string {
	var output []string
	for _, candidate := range c {
		output = append(output, candidate.String())
	}

	return strings.Join(output, ", ")
}

// ExtractContent returns relevant content.
func ExtractContent(page io.Reader) (string, error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", err
	}

	document.Find("script,style").Each(func(i int, s *goquery.Selection) {
		removeNodes(s)
	})

	transformMisusedDivsIntoParagraphs(document)
	removeUnlikelyCandidates(document)

	candidates := getCandidates(document)
	logger.Debug("[Readability] Candidates: %v", candidates)

	topCandidate := getTopCandidate(document, candidates)
	logger.Debug("[Readability] TopCandidate: %v", topCandidate)

	output := getArticle(topCandidate, candidates)
	return output, nil
}

// Now that we have the top candidate, look through its siblings for content that might also be related.
// Things like preambles, content split by ads that we removed, etc.
func getArticle(topCandidate *candidate, candidates candidateList) string {
	output := bytes.NewBufferString("<div>")
	siblingScoreThreshold := float32(math.Max(10, float64(topCandidate.score*.2)))

	topCandidate.selection.Siblings().Union(topCandidate.selection).Each(func(i int, s *goquery.Selection) {
		append := false
		node := s.Get(0)

		if node == topCandidate.Node() {
			append = true
		} else if c, ok := candidates[node]; ok && c.score >= siblingScoreThreshold {
			append = true
		}

		if s.Is("p") {
			linkDensity := getLinkDensity(s)
			content := s.Text()
			contentLength := len(content)

			if contentLength >= 80 && linkDensity < .25 {
				append = true
			} else if contentLength < 80 && linkDensity == 0 && sentenceRegexp.MatchString(content) {
				append = true
			}
		}

		if append {
			tag := "div"
			if s.Is("p") {
				tag = node.Data
			}

			html, _ := s.Html()
			fmt.Fprintf(output, "<%s>%s</%s>", tag, html, tag)
		}
	})

	output.Write([]byte("</div>"))
	return output.String()
}

func removeUnlikelyCandidates(document *goquery.Document) {
	document.Find("*").Not("html,body").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		str := class + id

		if blacklistCandidatesRegexp.MatchString(str) || (unlikelyCandidatesRegexp.MatchString(str) && !okMaybeItsACandidateRegexp.MatchString(str)) {
			removeNodes(s)
		}
	})
}

func getTopCandidate(document *goquery.Document, candidates candidateList) *candidate {
	var best *candidate

	for _, c := range candidates {
		if best == nil {
			best = c
		} else if best.score < c.score {
			best = c
		}
	}

	if best == nil {
		best = &candidate{document.Find("body"), 0}
	}

	return best
}

// Loop through all paragraphs, and assign a score to them based on how content-y they look.
// Then add their score to their parent node.
// A score is determined by things like number of commas, class names, etc.
// Maybe eventually link density.
func getCandidates(document *goquery.Document) candidateList {
	candidates := make(candidateList)

	document.Find(defaultTagsToScore).Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		// If this paragraph is less than 25 characters, don't even count it.
		if len(text) < 25 {
			return
		}

		parent := s.Parent()
		parentNode := parent.Get(0)

		grandParent := parent.Parent()
		var grandParentNode *html.Node
		if grandParent.Length() > 0 {
			grandParentNode = grandParent.Get(0)
		}

		if _, found := candidates[parentNode]; !found {
			candidates[parentNode] = scoreNode(parent)
		}

		if grandParentNode != nil {
			if _, found := candidates[grandParentNode]; !found {
				candidates[grandParentNode] = scoreNode(grandParent)
			}
		}

		// Add a point for the paragraph itself as a base.
		contentScore := float32(1.0)

		// Add points for any commas within this paragraph.
		contentScore += float32(strings.Count(text, ",") + 1)

		// For every 100 characters in this paragraph, add another point. Up to 3 points.
		contentScore += float32(math.Min(float64(int(len(text)/100.0)), 3))

		candidates[parentNode].score += contentScore
		if grandParentNode != nil {
			candidates[grandParentNode].score += contentScore / 2.0
		}
	})

	// Scale the final candidates score based on link density. Good content
	// should have a relatively small link density (5% or less) and be mostly
	// unaffected by this operation
	for _, candidate := range candidates {
		candidate.score = candidate.score * (1 - getLinkDensity(candidate.selection))
	}

	return candidates
}

func scoreNode(s *goquery.Selection) *candidate {
	c := &candidate{selection: s, score: 0}

	switch s.Get(0).DataAtom.String() {
	case "div":
		c.score += 5
	case "pre", "td", "blockquote", "img":
		c.score += 3
	case "address", "ol", "ul", "dl", "dd", "dt", "li", "form":
		c.score -= 3
	case "h1", "h2", "h3", "h4", "h5", "h6", "th":
		c.score -= 5
	}

	c.score += getClassWeight(s)
	return c
}

// Get the density of links as a percentage of the content
// This is the amount of text that is inside a link divided by the total text in the node.
func getLinkDensity(s *goquery.Selection) float32 {
	linkLength := len(s.Find("a").Text())
	textLength := len(s.Text())

	if textLength == 0 {
		return 0
	}

	return float32(linkLength) / float32(textLength)
}

// Get an elements class/id weight. Uses regular expressions to tell if this
// element looks good or bad.
func getClassWeight(s *goquery.Selection) float32 {
	weight := 0
	class, _ := s.Attr("class")
	id, _ := s.Attr("id")

	if class != "" {
		if negativeRegexp.MatchString(class) {
			weight -= 25
		}

		if positiveRegexp.MatchString(class) {
			weight += 25
		}
	}

	if id != "" {
		if negativeRegexp.MatchString(id) {
			weight -= 25
		}

		if positiveRegexp.MatchString(id) {
			weight += 25
		}
	}

	return float32(weight)
}

func transformMisusedDivsIntoParagraphs(document *goquery.Document) {
	document.Find("div").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		if !divToPElementsRegexp.MatchString(html) {
			node := s.Get(0)
			node.Data = "p"
		}
	})
}

func removeNodes(s *goquery.Selection) {
	s.Each(func(i int, s *goquery.Selection) {
		parent := s.Parent()
		if parent.Length() > 0 {
			parent.Get(0).RemoveChild(s.Get(0))
		}
	})
}
