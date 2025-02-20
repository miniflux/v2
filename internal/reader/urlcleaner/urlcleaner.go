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
// https://github.com/Smile4ever/Neat-URL/blob/master/data/default-params-by-category.json
// https://github.com/brave/brave-core/blob/master/components/query_filter/utils.cc
// https://developers.google.com/analytics/devguides/collection/ga4/reference/config
var trackingParams = map[string]bool{
	// Facebook Click Identifiers
	"fbclid":          true,
	"_openstat":       true,
	"fb_action_ids":   true,
	"fb_action_types": true,
	"fb_ref":          true,
	"fb_source":       true,
	"fb_comment_id":   true,

	// Google Click Identifiers
	"gclid":  true,
	"dclid":  true,
	"gbraid": true,
	"wbraid": true,
	"gclsrc": true,

	// Google Analytics
	"campaign_id":      true,
	"campaign_medium":  true,
	"campaign_name":    true,
	"campaign_source":  true,
	"campaign_term":    true,
	"campaign_content": true,

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
	"_hsmi":         true,
	"hsctatracking": true,

	// Olytics
	"rb_clickid":  true,
	"oly_anon_id": true,
	"oly_enc_id":  true,

	// Vero Click Identifier
	"vero_id":   true,
	"vero_conv": true,

	// Marketo email tracking
	"mkt_tok": true,

	// Adobe email tracking
	"sc_cid": true,

	// Beehiiv
	"_bhlid": true,

	// Branch.io
	"_branch_match_id": true,
	"_branch_referrer": true,
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
	hasTrackers := false

	// Remove tracking parameters
	for param := range queryParams {
		lowerParam := strings.ToLower(param)
		if trackingParams[lowerParam] || strings.HasPrefix(lowerParam, "utm_") {
			queryParams.Del(param)
			hasTrackers = true
		}
	}

	// Do not modify the URL if there are no tracking parameters
	if !hasTrackers {
		return inputURL, nil
	}

	parsedURL.RawQuery = queryParams.Encode()

	// Remove trailing "?" if query string is empty
	cleanedURL := parsedURL.String()
	cleanedURL = strings.TrimSuffix(cleanedURL, "?")

	return cleanedURL, nil
}
