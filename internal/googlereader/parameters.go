// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

const (
	// ParamItemIDs - name of the parameter with the item ids
	ParamItemIDs = "i"
	// ParamStreamID - name of the parameter containing the stream to be included
	ParamStreamID = "s"
	// ParamStreamExcludes - name of the parameter containing streams to be excluded
	ParamStreamExcludes = "xt"
	// ParamStreamFilters - name of the parameter containing streams to be included
	ParamStreamFilters = "it"
	// ParamStreamMaxItems - name of the parameter containing number of items per page/max items returned
	ParamStreamMaxItems = "n"
	// ParamStreamOrder - name of the parameter containing the sort criteria
	ParamStreamOrder = "r"
	// ParamStreamStartTime - name of the parameter containing epoch timestamp, filtering items older than
	ParamStreamStartTime = "ot"
	// ParamStreamStopTime - name of the parameter containing epoch timestamp, filtering items newer than
	ParamStreamStopTime = "nt"
	// ParamTagsRemove - name of the parameter containing tags (streams) to be removed
	ParamTagsRemove = "r"
	// ParamTagsAdd - name of the parameter containing tags (streams) to be added
	ParamTagsAdd = "a"
	// ParamSubscribeAction - name of the parameter indicating the action to take for subscription/edit
	ParamSubscribeAction = "ac"
	// ParamTitle - name of the parameter for the title of the subscription
	ParamTitle = "t"
	// ParamQuickAdd - name of the parameter for a URL being quick subscribed to
	ParamQuickAdd = "quickadd"
	// ParamDestination - name of the parameter for the new name of a tag
	ParamDestination = "dest"
	// ParamContinuation -  name of the parameter for callers to pass to receive the next page of results
	ParamContinuation = "c"
	// ParamStreamType - name of the parameter for unix timestamp
	ParamTimestamp = "ts"
)
