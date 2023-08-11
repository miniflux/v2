// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import (
	"fmt"
	"strconv"
	"strings"
)

type ImageCandidate struct {
	ImageURL   string
	Descriptor string
}

type ImageCandidates []*ImageCandidate

func (c ImageCandidates) String() string {
	var htmlCandidates []string

	for _, imageCandidate := range c {
		var htmlCandidate string
		if imageCandidate.Descriptor != "" {
			htmlCandidate = fmt.Sprintf(`%s %s`, imageCandidate.ImageURL, imageCandidate.Descriptor)
		} else {
			htmlCandidate = imageCandidate.ImageURL
		}

		htmlCandidates = append(htmlCandidates, htmlCandidate)
	}

	return strings.Join(htmlCandidates, ", ")
}

// ParseSrcSetAttribute returns the list of image candidates from the set.
// https://html.spec.whatwg.org/#parse-a-srcset-attribute
func ParseSrcSetAttribute(attributeValue string) (imageCandidates ImageCandidates) {
	unparsedCandidates := strings.Split(attributeValue, ", ")

	for _, unparsedCandidate := range unparsedCandidates {
		if candidate, err := parseImageCandidate(unparsedCandidate); err == nil {
			imageCandidates = append(imageCandidates, candidate)
		}
	}

	return imageCandidates
}

func parseImageCandidate(input string) (*ImageCandidate, error) {
	input = strings.TrimSpace(input)
	parts := strings.Split(strings.TrimSpace(input), " ")
	nbParts := len(parts)

	if nbParts > 2 || nbParts == 0 {
		return nil, fmt.Errorf(`srcset: invalid number of descriptors`)
	}

	if nbParts == 2 {
		if !isValidWidthOrDensityDescriptor(parts[1]) {
			return nil, fmt.Errorf(`srcset: invalid descriptor`)
		}
		return &ImageCandidate{ImageURL: parts[0], Descriptor: parts[1]}, nil
	}

	return &ImageCandidate{ImageURL: parts[0]}, nil
}

func isValidWidthOrDensityDescriptor(value string) bool {
	if value == "" {
		return false
	}

	lastChar := value[len(value)-1:]
	if lastChar != "w" && lastChar != "x" {
		return false
	}

	_, err := strconv.ParseFloat(value[0:len(value)-1], 32)
	return err == nil
}
