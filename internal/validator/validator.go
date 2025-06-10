// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var domainRegex = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

// ValidateRange makes sure the offset/limit values are valid.
func ValidateRange(offset, limit int) error {
	if offset < 0 {
		return fmt.Errorf(`offset value should be >= 0`)
	}

	if limit < 0 {
		return fmt.Errorf(`limit value should be >= 0`)
	}

	return nil
}

// ValidateDirection makes sure the sorting direction is valid.
func ValidateDirection(direction string) error {
	switch direction {
	case "asc", "desc":
		return nil
	}

	return fmt.Errorf(`invalid direction, valid direction values are: "asc" or "desc"`)
}

// IsValidRegex verifies if the regex can be compiled.
func IsValidRegex(expr string) bool {
	_, err := regexp.Compile(expr)
	return err == nil
}

// IsValidURL verifies if the provided value is a valid absolute URL.
func IsValidURL(absoluteURL string) bool {
	_, err := url.ParseRequestURI(absoluteURL)
	return err == nil
}

func IsValidDomain(domain string) bool {
	domain = strings.ToLower(domain)

	if len(domain) < 1 || len(domain) > 253 {
		return false
	}

	return domainRegex.MatchString(domain)
}

func IsValidDomainList(value string) bool {
	domains := strings.Split(strings.TrimSpace(value), " ")
	for _, domain := range domains {
		if !IsValidDomain(domain) {
			return false
		}
	}

	return true
}
