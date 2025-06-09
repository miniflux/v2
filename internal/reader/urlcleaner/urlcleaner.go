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

// Outbound tracking parameters are appending the website's url to outbound links.
var trackingParamsOutbound = map[string]bool{
	// Ghost
	"ref": true,
}

func RemoveTrackingParameters(parsedFeedURL, parsedSiteURL, parsedInputUrl *url.URL) (string, error) {
	if parsedFeedURL == nil || parsedSiteURL == nil || parsedInputUrl == nil {
		return "", fmt.Errorf("urlcleaner: one of the URLs is nil")
	}

	queryParams := parsedInputUrl.Query()
	hasTrackers := false

	// Remove tracking parameters
	for param := range queryParams {
		lowerParam := strings.ToLower(param)
		if trackingParams[lowerParam] || strings.HasPrefix(lowerParam, "utm_") {
			queryParams.Del(param)
			hasTrackers = true
		}
		if trackingParamsOutbound[lowerParam] {
			// handle duplicate parameters like ?a=b&a=c&a=d…
			for _, value := range queryParams[param] {
				if value == parsedFeedURL.Hostname() || value == parsedSiteURL.Hostname() {
					queryParams.Del(param)
					hasTrackers = true
					break
				}
			}
		}
	}

	// Do not modify the URL if there are no tracking parameters
	if !hasTrackers {
		return parsedInputUrl.String(), nil
	}

	parsedInputUrl.RawQuery = queryParams.Encode()
	cleanedURL := strings.TrimSuffix(parsedInputUrl.String(), "?")

	return cleanedURL, nil
}
