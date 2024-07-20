// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package urlcleaner // import "miniflux.app/v2/internal/reader/urlcleaner"

import (
	"fmt"
	"net/url"
	"strings"
)

// Interesting lists:
// https://raw.githubusercontent.com/AdguardTeam/AdguardFilters/master/TrackParamFilter/sections/general_url.txt
// https://firefox.settings.services.mozilla.com/v1/buckets/main/collections/query-stripping/records
var trackingParams = map[string]bool{
	// https://en.wikipedia.org/wiki/UTM_parameters#Parameters
	"utm_source":   true,
	"utm_medium":   true,
	"utm_campaign": true,
	"utm_term":     true,
	"utm_content":  true,

	// Facebook Click Identifiers
	"fbclid":    true,
	"_openstat": true,

	// Google Click Identifiers
	"gclid":  true,
	"dclid":  true,
	"gbraid": true,
	"wbraid": true,

	// Yandex Click Identifiers
	"yclid":  true,
	"ysclid": true,

	// Twitter Click Identifier
	"twclid": true,

	// Microsoft Click Identifier
	"msclkid": true,

	// Mailchimp Click Identifiers
	"mc_cid": true,
	"mc_eid": true,

	// Wicked Reports click tracking
	"wickedid": true,

	// Hubspot Click Identifiers
	"hsa_cam":       true,
	"_hsenc":        true,
	"__hssc":        true,
	"__hstc":        true,
	"__hsfp":        true,
	"hsctatracking": true,

	// Olytics
	"rb_clickid":  true,
	"oly_anon_id": true,
	"oly_enc_id":  true,

	// Vero Click Identifier
	"vero_id": true,

	// Marketo email tracking
	"mkt_tok": true,
}

func RemoveTrackingParameters(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("urlcleaner: error parsing URL: %v", err)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return inputURL, nil
	}

	queryParams := parsedURL.Query()

	// Remove tracking parameters
	for param := range queryParams {
		if trackingParams[strings.ToLower(param)] {
			queryParams.Del(param)
		}
	}

	parsedURL.RawQuery = queryParams.Encode()

	// Remove trailing "?" if query string is empty
	cleanedURL := parsedURL.String()
	cleanedURL = strings.TrimSuffix(cleanedURL, "?")

	return cleanedURL, nil
}
