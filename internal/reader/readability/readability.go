// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/v2/internal/reader/readability"

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const defaultTagsToScore = "section,h2,h3,h4,h5,h6,p,td,pre,div"

var (
	strongCandidates  = [...]string{"popupbody", "-ad", "g-plus"}
	maybeCandidate    = [...]string{"and", "article", "body", "column", "main", "shadow"}
	unlikelyCandidate = [...]string{"banner", "breadcrumbs", "combx", "comment", "community", "cover-wrap", "disqus", "extra", "foot", "header", "legends", "menu", "modal", "related", "remark", "replies", "rss", "shoutbox", "sidebar", "skyscraper", "social", "sponsor", "supplemental", "ad-break", "agegate", "pagination", "pager", "popup", "yom-remote"}

	positiveKeywords = [...]string{"article", "blog", "body", "content", "entry", "h-entry", "hentry", "main", "page", "pagination", "post", "story", "text"}
	negativeKeywords = [...]string{"author", "banner", "byline", "com-", "combx", "comment", "contact", "dateline", "foot", "hid", "masthead", "media", "meta", "modal", "outbrain", "promo", "related", "scroll", "share", "shopping", "shoutbox", "sidebar", "skyscraper", "sponsor", "tags", "tool", "widget", "writtenby"}
)

type candidate struct {
	selection *goquery.Selection
	score     float32
}

func (c *candidate) Node() *html.Node {
	if c.selection.Length() == 0 {
		return nil
	}
	return c.selection.Get(0)
}

func (c *candidate) String() string {
	node := c.Node()
	if node == nil {
		return fmt.Sprintf("empty => %f", c.score)
	}

	id, _ := c.selection.Attr("id")
	class, _ := c.selection.Attr("class")

	switch {
	case id != "" && class != "":
		return fmt.Sprintf("%s#%s.%s => %f", node.DataAtom, id, class, c.score)
	case id != "":
		return fmt.Sprintf("%s#%s => %f", node.DataAtom, id, c.score)
	case class != "":
		return fmt.Sprintf("%s.%s => %f", node.DataAtom, class, c.score)
	}

	return fmt.Sprintf("%s => %f", node.DataAtom, c.score)
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
func ExtractContent(page io.Reader) (baseURL string, extractedContent string, err error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", "", err
	}

	if hrefValue, exists := document.FindMatcher(goquery.Single("head base")).Attr("href"); exists {
		hrefValue = strings.TrimSpace(hrefValue)
		if urllib.IsAbsoluteURL(hrefValue) {
			baseURL = hrefValue
		}
	}

	document.Find("script,style").Remove()

	transformMisusedDivsIntoParagraphs(document)
	removeUnlikelyCandidates(document)

	candidates := getCandidates(document)
	topCandidate := getTopCandidate(document, candidates)

	slog.Debug("Readability parsing",
		slog.String("base_url", baseURL),
		slog.Any("candidates", candidates),
		slog.Any("topCandidate", topCandidate),
	)

	extractedContent = getArticle(topCandidate, candidates)
	return baseURL, extractedContent, nil
}

func getSelectionLength(s *goquery.Selection) int {
	var getLengthOfTextContent func(*html.Node) int
	getLengthOfTextContent = func(n *html.Node) int {
		total := 0
		if n.Type == html.TextNode {
			total += len(n.Data)
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				total += getLengthOfTextContent(c)
			}
		}
		return total
	}

	sum := 0
	for _, n := range s.Nodes {
		sum += getLengthOfTextContent(n)
	}
	return sum
}

// Now that we have the top candidate, look through its siblings for content that might also be related.
// Things like preambles, content split by ads that we removed, etc.
func getArticle(topCandidate *candidate, candidates candidateList) string {
	var output strings.Builder
	output.WriteString("<div>")
	siblingScoreThreshold := max(10, topCandidate.score/5)

	topCandidate.selection.Siblings().Union(topCandidate.selection).Each(func(i int, s *goquery.Selection) {
		append := false
		tag := "div"
		node := s.Get(0)

		topNode := topCandidate.Node()
		if topNode != nil && node == topNode {
			append = true
		} else if c, ok := candidates[node]; ok && c.score >= siblingScoreThreshold {
			append = true
		} else if s.Is("p") {
			tag = node.Data
			linkDensity := getLinkDensity(s)
			contentLength := getSelectionLength(s)

			if contentLength >= 80 {
				if linkDensity < .25 {
					append = true
				}
			} else {
				if linkDensity == 0 {
					// It's a small selection, so .Text doesn't impact performances too much.
					content := s.Text()
					if containsSentence(content) {
						append = true
					}
				}
			}
		}

		if append {
			html, _ := s.Html()
			output.WriteString("<" + tag + ">" + html + "</" + tag + ">")
		}
	})

	output.WriteString("</div>")
	return output.String()
}
func shouldRemoveCandidate(str string) bool {
	str = strings.ToLower(str)

	// Those candidates have no false-positives, no need to check against `maybeCandidate`
	for _, strongCandidate := range strongCandidates {
		if strings.Contains(str, strongCandidate) {
			return true
		}
	}

	for _, unlikelyCandidate := range unlikelyCandidate {
		if strings.Contains(str, unlikelyCandidate) {
			// Do we have a false positive?
			for _, maybe := range maybeCandidate {
				if strings.Contains(str, maybe) {
					return false
				}
			}

			// Nope, it's a true positive!
			return true
		}
	}
	return false
}

func removeUnlikelyCandidates(document *goquery.Document) {
	document.Find("*").Each(func(i int, s *goquery.Selection) {
		if s.Length() == 0 || s.Get(0).Data == "html" || s.Get(0).Data == "body" {
			return
		}

		// Don't remove elements within code blocks (pre or code tags)
		if s.Closest("pre, code").Length() > 0 {
			return
		}

		if class, ok := s.Attr("class"); ok && shouldRemoveCandidate(class) {
			s.Remove()
		} else if id, ok := s.Attr("id"); ok && shouldRemoveCandidate(id) {
			s.Remove()
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
func getCandidates(document *goquery.Document) candidateList {
	candidates := make(candidateList)

	document.Find(defaultTagsToScore).Each(func(i int, s *goquery.Selection) {
		textLen := getSelectionLength(s)

		// If this paragraph is less than 25 characters, don't even count it.
		if textLen < 25 {
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
		text := s.Text()
		contentScore += float32(strings.Count(text, ",") + 1)

		// For every 100 characters in this paragraph, add another point. Up to 3 points.
		contentScore += float32(min(textLen/100.0, 3))

		candidates[parentNode].score += contentScore
		if grandParentNode != nil {
			candidates[grandParentNode].score += contentScore / 2.0
		}
	})

	// Scale the final candidates score based on link density. Good content
	// should have a relatively small link density (5% or less) and be mostly
	// unaffected by this operation
	for _, candidate := range candidates {
		candidate.score *= (1 - getLinkDensity(candidate.selection))
	}

	return candidates
}

func scoreNode(s *goquery.Selection) *candidate {
	c := &candidate{selection: s, score: 0}

	// Check if selection is empty to avoid panic
	if s.Length() == 0 {
		return c
	}

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
	sum := getSelectionLength(s)
	if sum == 0 {
		return 0
	}

	linkLength := getSelectionLength(s.Find("a"))

	return float32(linkLength) / float32(sum)
}

// Get an elements class/id weight. Uses regular expressions to tell if this
// element looks good or bad.
func getClassWeight(s *goquery.Selection) float32 {
	weight := 0

	if class, ok := s.Attr("class"); ok {
		weight += getWeight(class)
	}
	if id, ok := s.Attr("id"); ok {
		weight += getWeight(id)
	}

	return float32(weight)
}

func getWeight(s string) int {
	s = strings.ToLower(s)
	for _, keyword := range negativeKeywords {
		if strings.Contains(s, keyword) {
			return -25
		}
	}
	for _, keyword := range positiveKeywords {
		if strings.Contains(s, keyword) {
			return +25
		}
	}
	return 0
}

func transformMisusedDivsIntoParagraphs(document *goquery.Document) {
	document.Find("div").Each(func(i int, s *goquery.Selection) {
		nodes := s.Children().Nodes

		if len(nodes) == 0 {
			node := s.Get(0)
			node.Data = "p"
			return
		}

		for _, node := range nodes {
			switch node.Data {
			case "a", "blockquote", "div", "dl",
				"img", "ol", "p", "pre",
				"table", "ul":
				return
			default:
				currentNode := s.Get(0)
				currentNode.Data = "p"
			}
		}
	})
}

func containsSentence(content string) bool {
	return strings.HasSuffix(content, ".") || strings.Contains(content, ". ")
}
