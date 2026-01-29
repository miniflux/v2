// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import (
	"math"
	"strconv"
	"strings"
)

type imageCandidate struct {
	ImageURL   string
	Descriptor string
}

type imageCandidates []*imageCandidate

func (c imageCandidates) String() string {
	htmlCandidates := make([]string, 0, len(c))

	for _, imageCandidate := range c {
		var htmlCandidate string
		if imageCandidate.Descriptor != "" {
			htmlCandidate = imageCandidate.ImageURL + " " + imageCandidate.Descriptor
		} else {
			htmlCandidate = imageCandidate.ImageURL
		}

		htmlCandidates = append(htmlCandidates, htmlCandidate)
	}

	return strings.Join(htmlCandidates, ", ")
}

// ParseSrcSetAttribute returns the list of image candidates from the set.
// Parsing behavior follows the WebKit HTMLSrcsetParser implementation.
// https://html.spec.whatwg.org/#parse-a-srcset-attribute
func ParseSrcSetAttribute(attributeValue string) (candidates imageCandidates) {
	if attributeValue == "" {
		return nil
	}

	var position uint = 0
	for position < uint(len(attributeValue)) {
		position = skipWhileHTMLSpaceOrComma(attributeValue, position)
		if position >= uint(len(attributeValue)) {
			break
		}

		urlStart := position
		position = skipUntilASCIIWhitespace(attributeValue, position)
		imageURL := attributeValue[urlStart:position]
		if imageURL == "" {
			continue
		}

		var result descriptorParsingResult
		if imageURL[len(imageURL)-1] == ',' {
			imageURL = strings.TrimRight(imageURL, ",")
			if imageURL == "" {
				continue
			}
		} else {
			position = skipWhileASCIIWhitespace(attributeValue, position)
			descriptorTokens, newPosition := tokenizeDescriptors(attributeValue, position)
			position = newPosition
			if !parseDescriptors(descriptorTokens, &result) {
				continue
			}
		}

		candidates = append(candidates, &imageCandidate{
			ImageURL:   imageURL,
			Descriptor: serializeDescriptor(result),
		})
	}

	return candidates
}

type descriptorParsingResult struct {
	density        float64
	resourceWidth  uint
	resourceHeight uint
	hasDensity     bool
	hasWidth       bool
	hasHeight      bool
}

func (r *descriptorParsingResult) setDensity(value float64) {
	r.density = value
	r.hasDensity = true
}

func (r *descriptorParsingResult) setResourceWidth(value uint) {
	r.resourceWidth = value
	r.hasWidth = true
}

func (r *descriptorParsingResult) setResourceHeight(value uint) {
	r.resourceHeight = value
	r.hasHeight = true
}

func serializeDescriptor(result descriptorParsingResult) string {
	if result.hasDensity {
		return formatFloat(result.density) + "x"
	}
	if result.hasWidth {
		return strconv.FormatUint(uint64(result.resourceWidth), 10) + "w"
	}
	return ""
}

func parseDescriptors(descriptors []string, result *descriptorParsingResult) bool {
	for _, descriptor := range descriptors {
		if descriptor == "" {
			continue
		}
		lastIndex := len(descriptor) - 1
		descriptorChar := descriptor[lastIndex]
		value := descriptor[:lastIndex]

		switch descriptorChar {
		case 'x':
			if result.hasDensity || result.hasHeight || result.hasWidth {
				return false
			}
			density, ok := parseValidHTMLFloatingPointNumber(value)
			if !ok || density < 0 {
				return false
			}
			result.setDensity(density)
		case 'w':
			if result.hasDensity || result.hasWidth {
				return false
			}
			width, ok := parseValidHTMLNonNegativeInteger(value)
			if !ok || width <= 0 {
				return false
			}
			result.setResourceWidth(width)
		case 'h':
			if result.hasDensity || result.hasHeight {
				return false
			}
			height, ok := parseValidHTMLNonNegativeInteger(value)
			if !ok || height <= 0 {
				return false
			}
			result.setResourceHeight(height)
		default:
			return false
		}
	}

	return !result.hasHeight || result.hasWidth
}

type descriptorTokenizerState int

const (
	descriptorStateInitial descriptorTokenizerState = iota
	descriptorStateInParenthesis
	descriptorStateAfterToken
)

func tokenizeDescriptors(input string, start uint) (tokens []string, newPosition uint) {
	state := descriptorStateInitial
	currentStart := start
	currentSet := true
	position := start

	appendDescriptorAndReset := func(position uint) {
		if currentSet && position > currentStart {
			tokens = append(tokens, input[currentStart:position])
		}
		currentSet = false
	}

	appendCharacter := func(position uint) {
		if !currentSet {
			currentStart = position
			currentSet = true
		}
	}

	for {
		if position >= uint(len(input)) {
			if state != descriptorStateAfterToken {
				appendDescriptorAndReset(position)
			}
			return tokens, position
		}

		character := input[position]
		switch state {
		case descriptorStateInitial:
			switch {
			case isComma(character):
				appendDescriptorAndReset(position)
				position++
				return tokens, position
			case isASCIIWhitespace(character):
				appendDescriptorAndReset(position)
				currentStart = position + 1
				currentSet = true
				state = descriptorStateAfterToken
			case character == '(':
				appendCharacter(position)
				state = descriptorStateInParenthesis
			default:
				appendCharacter(position)
			}
		case descriptorStateInParenthesis:
			if character == ')' {
				appendCharacter(position)
				state = descriptorStateInitial
			} else {
				appendCharacter(position)
			}
		case descriptorStateAfterToken:
			if !isASCIIWhitespace(character) {
				state = descriptorStateInitial
				currentStart = position
				currentSet = true
				position--
			}
		}

		position++
	}
}

func parseValidHTMLNonNegativeInteger(value string) (uint, bool) {
	if value == "" {
		return 0, false
	}

	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0, false
		}
	}

	parsed, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return 0, false
	}

	return uint(parsed), true
}

func parseValidHTMLFloatingPointNumber(value string) (float64, bool) {
	if value == "" {
		return 0, false
	}
	if value[0] == '+' || value[len(value)-1] == '.' {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}

	return parsed, true
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func skipWhileHTMLSpaceOrComma(value string, position uint) uint {
	for position < uint(len(value)) && (isASCIIWhitespace(value[position]) || isComma(value[position])) {
		position++
	}
	return position
}

func skipWhileASCIIWhitespace(value string, position uint) uint {
	for position < uint(len(value)) && isASCIIWhitespace(value[position]) {
		position++
	}
	return position
}

func skipUntilASCIIWhitespace(value string, position uint) uint {
	for position < uint(len(value)) && !isASCIIWhitespace(value[position]) {
		position++
	}
	return position
}

func isASCIIWhitespace(character byte) bool {
	switch character {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}

func isComma(character byte) bool {
	return character == ','
}
