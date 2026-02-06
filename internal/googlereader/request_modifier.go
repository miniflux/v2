// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"fmt"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
)

type requestModifiers struct {
	ExcludeTargets    []Stream
	FilterTargets     []Stream
	Streams           []Stream
	Count             int
	Offset            int
	SortDirection     string
	StartTime         int64
	StopTime          int64
	ContinuationToken string
	UserID            int64
}

func (r requestModifiers) String() string {
	var results []string

	results = append(results, fmt.Sprintf("UserID: %d", r.UserID))

	streamStr := make([]string, 0, len(r.Streams))
	for _, s := range r.Streams {
		streamStr = append(streamStr, s.String())
	}
	results = append(results, fmt.Sprintf("Streams: [%s]", strings.Join(streamStr, ", ")))

	exclusions := make([]string, 0, len(r.ExcludeTargets))
	for _, s := range r.ExcludeTargets {
		exclusions = append(exclusions, s.String())
	}
	results = append(results, fmt.Sprintf("Exclusions: [%s]", strings.Join(exclusions, ", ")))

	filters := make([]string, 0, len(r.FilterTargets))
	for _, s := range r.FilterTargets {
		filters = append(filters, s.String())
	}
	results = append(results, fmt.Sprintf("Filters: [%s]", strings.Join(filters, ", ")))

	results = append(results, fmt.Sprintf("Count: %d", r.Count))
	results = append(results, fmt.Sprintf("Offset: %d", r.Offset))
	results = append(results, "Sort Direction: "+r.SortDirection)
	results = append(results, "Continuation Token: "+r.ContinuationToken)
	results = append(results, fmt.Sprintf("Start Time: %d", r.StartTime))
	results = append(results, fmt.Sprintf("Stop Time: %d", r.StopTime))

	return strings.Join(results, "; ")
}

func parseStreamFilterFromRequest(r *http.Request) (requestModifiers, error) {
	userID := request.UserID(r)
	result := requestModifiers{
		SortDirection: "desc",
		UserID:        userID,
	}

	streamOrder := request.QueryStringParam(r, paramStreamOrder, "d")
	if streamOrder == "o" {
		result.SortDirection = "asc"
	}
	var err error
	result.Streams, err = getStreams(request.QueryStringParamList(r, paramStreamID), userID)
	if err != nil {
		return requestModifiers{}, err
	}
	result.ExcludeTargets, err = getStreams(request.QueryStringParamList(r, paramStreamExcludes), userID)
	if err != nil {
		return requestModifiers{}, err
	}

	result.FilterTargets, err = getStreams(request.QueryStringParamList(r, paramStreamFilters), userID)
	if err != nil {
		return requestModifiers{}, err
	}

	result.Count = request.QueryIntParam(r, paramStreamMaxItems, 0)
	result.Offset = request.QueryIntParam(r, paramContinuation, 0)
	result.StartTime = request.QueryInt64Param(r, paramStreamStartTime, int64(0))
	result.StopTime = request.QueryInt64Param(r, paramStreamStopTime, int64(0))
	return result, nil
}
