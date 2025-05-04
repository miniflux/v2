// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	ItemIDPrefix = "tag:google.com,2005:reader/item/"
)

func convertEntryIDToLongFormItemID(entryID int64) string {
	// The entry ID is a 64-bit integer, so we need to format it as a 16-character hexadecimal string.
	return ItemIDPrefix + fmt.Sprintf("%016x", entryID)
}

// parseItemID parses a Google Reader ID string.
// It supports both the long form (tag:google.com,2005:reader/item/<hex_id>) and the short form (<decimal_id>).
// It returns the parsed ID as a int64 and an error if parsing fails.
func parseItemID(itemIDValue string) (int64, error) {
	if strings.HasPrefix(itemIDValue, ItemIDPrefix) {
		hexID := strings.TrimPrefix(itemIDValue, ItemIDPrefix)

		// It's always 16 characters wide.
		if len(hexID) != 16 {
			return 0, errors.New("long form ID has incorrect length")
		}

		parsedID, err := strconv.ParseInt(hexID, 16, 64)
		if err != nil {
			return 0, errors.New("failed to parse long form hex ID: " + err.Error())
		}
		return parsedID, nil
	} else {
		parsedID, err := strconv.ParseInt(itemIDValue, 10, 64)
		if err != nil {
			return 0, errors.New("failed to parse short form decimal ID: " + err.Error())
		}
		return parsedID, nil
	}
}

func parseItemIDsFromRequest(r *http.Request) ([]int64, error) {
	items := r.Form[ParamItemIDs]
	if len(items) == 0 {
		return nil, fmt.Errorf("googlereader: no items requested")
	}

	itemIDs := make([]int64, len(items))
	for i, item := range items {
		itemID, err := parseItemID(item)
		if err != nil {
			return nil, fmt.Errorf("googlereader: failed to parse item ID %s: %w", item, err)
		}
		itemIDs[i] = itemID
	}

	return itemIDs, nil
}
