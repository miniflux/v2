// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	ItemIDPrefix = "tag:google.com,2005:reader/item/"
	ItemIDFormat = "tag:google.com,2005:reader/item/%016x"
)

func convertEntryIDToLongFormItemID(entryID int64) string {
	// The entry ID is a 64-bit integer, so we need to format it as a 16-character hexadecimal string.
	return fmt.Sprintf(ItemIDFormat, entryID)
}

// Expected format: "tag:google.com,2005:reader/item/00000000148b9369" (hexadecimal string with prefix and padding)
// NetNewsWire uses this format: "tag:google.com,2005:reader/item/2f2" (hexadecimal string with prefix and no padding)
// Reeder uses this format: "000000000000048c" (hexadecimal string without prefix and padding)
// Liferea uses this format: "12345" (decimal string)
// It returns the parsed ID as a int64 and an error if parsing fails.
func parseItemID(itemIDValue string) (int64, error) {
	var itemID int64
	if strings.HasPrefix(itemIDValue, ItemIDPrefix) {
		n, err := fmt.Sscanf(itemIDValue, ItemIDFormat, &itemID)
		if err != nil {
			return 0, fmt.Errorf("failed to parse hexadecimal item ID %s: %w", itemIDValue, err)
		}
		if n != 1 {
			return 0, fmt.Errorf("failed to parse hexadecimal item ID %s: expected 1 value, got %d", itemIDValue, n)
		}
		if itemID == 0 {
			return 0, fmt.Errorf("failed to parse hexadecimal item ID %s: item ID is zero", itemIDValue)
		}
		return itemID, nil
	}

	if len(itemIDValue) == 16 {
		if n, err := fmt.Sscanf(itemIDValue, "%016x", &itemID); err == nil && n == 1 {
			return itemID, nil
		}
	}

	itemID, err := strconv.ParseInt(itemIDValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decimal item ID %s: %w", itemIDValue, err)
	}

	return itemID, nil
}

func parseItemIDsFromRequest(r *http.Request) ([]int64, error) {
	items := r.Form[paramItemIDs]
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
