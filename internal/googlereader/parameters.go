// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

const (
	// paramItemIDs - name of the parameter with the item ids
	paramItemIDs = "i"
	// paramStreamID - name of the parameter containing the stream to be included
	paramStreamID = "s"
	// paramStreamExcludes - name of the parameter containing streams to be excluded
	paramStreamExcludes = "xt"
	// paramStreamFilters - name of the parameter containing streams to be included
	paramStreamFilters = "it"
	// paramStreamMaxItems - name of the parameter containing number of items per page/max items returned
	paramStreamMaxItems = "n"
	// paramStreamOrder - name of the parameter containing the sort criteria
	paramStreamOrder = "r"
	// paramStreamStartTime - name of the parameter containing epoch timestamp, filtering items older than
	paramStreamStartTime = "ot"
	// paramStreamStopTime - name of the parameter containing epoch timestamp, filtering items newer than
	paramStreamStopTime = "nt"
	// paramTagsRemove - name of the parameter containing tags (streams) to be removed
	paramTagsRemove = "r"
	// paramTagsAdd - name of the parameter containing tags (streams) to be added
	paramTagsAdd = "a"
	// paramSubscribeAction - name of the parameter indicating the action to take for subscription/edit
	paramSubscribeAction = "ac"
	// paramTitle - name of the parameter for the title of the subscription
	paramTitle = "t"
	// paramQuickAdd - name of the parameter for a URL being quick subscribed to
	paramQuickAdd = "quickadd"
	// paramDestination - name of the parameter for the new name of a tag
	paramDestination = "dest"
	// paramContinuation -  name of the parameter for callers to pass to receive the next page of results
	paramContinuation = "c"
	// paramTimestamp - name of the parameter for unix timestamp
	paramTimestamp = "ts"
)
