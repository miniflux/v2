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
	strongCandidatesToRemove  = [...]string{"popupbody", "-ad", "g-plus"}
	maybeCandidateToRemove    = [...]string{"and", "article", "body", "column", "main", "shadow", "content"}
	unlikelyCandidateToRemove = [...]string{"banner", "breadcrumbs", "combx", "comment", "community", "cover-wrap", "disqus", "extra", "foot", "header", "legends", "menu", "modal", "related", "remark", "replies", "rss", "shoutbox", "sidebar", "skyscraper", "social", "sponsor", "supplemental", "ad-break", "agegate", "pagination", "pager", "popup", "yom-remote"}

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

	removeUnlikelyCandidates(document)
	transformMisusedDivsIntoParagraphs(document)

	candidates := getCandidates(document)
	topCandidate := getTopCandidate(document, candidates)

	slog.Debug("Readability parsing",
		slog.String("base_url", baseURL),
		slog.String("candidates", candidates.String()),
		slog.String("topCandidate", topCandidate.String()),
	)

	extractedContent = getArticle(topCandidate, candidates)
	return baseURL, extractedContent, nil
}

func getSelectionLength(s *goquery.Selection) int {
	return sumMapOnSelection(s, func(s string) int { return len(s) })
}

func getSelectionCommaCount(s *goquery.Selection) int {
	return sumMapOnSelection(s, func(s string) int { return strings.Count(s, ",") })
}

// sumMapOnSelection maps `f` on the selection, and return the sum of the result.
// This construct is used instead of goquery.Selection's .Text() method,
// to avoid materializing the text to simply map/sum on it, saving a significant
// amount of memory of large selections, and reducing the pressure on the garbage-collector.
func sumMapOnSelection(s *goquery.Selection, f func(str string) int) int {
	var recursiveFunction func(*html.Node) int
	recursiveFunction = func(n *html.Node) int {
		total := 0
		if n.Type == html.TextNode {
			total += f(n.Data)
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				total += recursiveFunction(c)
			}
		}
		return total
	}

	sum := 0
	for _, n := range s.Nodes {
		sum += recursiveFunction(n)
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
					if containsSentence(s.Text()) {
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
	for _, strongCandidateToRemove := range strongCandidatesToRemove {
		if strings.Contains(str, strongCandidateToRemove) {
			return true
		}
	}

	for _, unlikelyCandidateToRemove := range unlikelyCandidateToRemove {
		if strings.Contains(str, unlikelyCandidateToRemove) {
			// Do we have a false positive?
			for _, maybeCandidateToRemove := range maybeCandidateToRemove {
				if strings.Contains(str, maybeCandidateToRemove) {
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
	// Only select tags with either a class or an id attribute,
	// and never the html nor body tags, as we don't want to ever remove them.
	selector := "[class]:not(body,html)" + "," + "[id]:not(body,html)"

	for _, s := range document.Find(selector).EachIter() {
		if s.Length() == 0 {
			continue
		}

		// Don't remove elements within code blocks (pre or code tags)
		if s.Closest("pre,code").Length() > 0 {
			continue
		}

		if class, ok := s.Attr("class"); ok && shouldRemoveCandidate(class) {
			s.Remove()
		} else if id, ok := s.Attr("id"); ok && shouldRemoveCandidate(id) {
			s.Remove()
		}
	}
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

		// Add a point for the paragraph itself as a base.
		contentScore := 1

		// Add points for any commas within this paragraph.
		contentScore += getSelectionCommaCount(s) + 1

		// For every 100 characters in this paragraph, add another point. Up to 3 points.
		contentScore += min(textLen/100, 3)

		parent := s.Parent()
		parentNode := parent.Get(0)
		if _, found := candidates[parentNode]; !found {
			candidates[parentNode] = scoreNode(parent)
		}
		candidates[parentNode].score += float32(contentScore)

		// The score of the current node influences its grandparent's one as well, but scaled to 50%.
		grandParent := parent.Parent()
		if grandParent.Length() > 0 {
			grandParentNode := grandParent.Get(0)
			if _, found := candidates[grandParentNode]; !found {
				candidates[grandParentNode] = scoreNode(grandParent)
			}
			candidates[grandParentNode].score += float32(contentScore) / 2.0
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

	switch s.Get(0).Data {
	case "div":
		c.score += 5
	case "pre", "td", "blockquote", "img":
		c.score += 3
	case "address", "ol", "ul", "dl", "dd", "dt", "li", "form":
		c.score -= 3
	case "h1", "h2", "h3", "h4", "h5", "h6", "th":
		c.score -= 5
	}

	if class, ok := s.Attr("class"); ok {
		c.score += getWeight(class)
	}
	if id, ok := s.Attr("id"); ok {
		c.score += getWeight(id)
	}

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

func getWeight(s string) float32 {
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
			s.Nodes[0].Data = "p"
			return
		}

		for _, node := range nodes {
			switch node.Data {
			case "a", "blockquote", "div", "dl",
				"img", "ol", "p", "pre",
				"table", "ul":
				return
			default:
				s.Nodes[0].Data = "p"
			}
		}
	})
}

func containsSentence(content string) bool {
	return strings.HasSuffix(content, ".") || strings.Contains(content, ". ")
}
