// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"fmt"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
)

type RequestModifiers struct {
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

func (r RequestModifiers) String() string {
	var results []string

	results = append(results, fmt.Sprintf("UserID: %d", r.UserID))

	var streamStr []string
	for _, s := range r.Streams {
		streamStr = append(streamStr, s.String())
	}
	results = append(results, fmt.Sprintf("Streams: [%s]", strings.Join(streamStr, ", ")))

	var exclusions []string
	for _, s := range r.ExcludeTargets {
		exclusions = append(exclusions, s.String())
	}
	results = append(results, fmt.Sprintf("Exclusions: [%s]", strings.Join(exclusions, ", ")))

	var filters []string
	for _, s := range r.FilterTargets {
		filters = append(filters, s.String())
	}
	results = append(results, fmt.Sprintf("Filters: [%s]", strings.Join(filters, ", ")))

	results = append(results, fmt.Sprintf("Count: %d", r.Count))
	results = append(results, fmt.Sprintf("Offset: %d", r.Offset))
	results = append(results, fmt.Sprintf("Sort Direction: %s", r.SortDirection))
	results = append(results, fmt.Sprintf("Continuation Token: %s", r.ContinuationToken))
	results = append(results, fmt.Sprintf("Start Time: %d", r.StartTime))
	results = append(results, fmt.Sprintf("Stop Time: %d", r.StopTime))

	return strings.Join(results, "; ")
}

func parseStreamFilterFromRequest(r *http.Request) (RequestModifiers, error) {
	userID := request.UserID(r)
	result := RequestModifiers{
		SortDirection: "desc",
		UserID:        userID,
	}

	streamOrder := request.QueryStringParam(r, ParamStreamOrder, "d")
	if streamOrder == "o" {
		result.SortDirection = "asc"
	}
	var err error
	result.Streams, err = getStreams(request.QueryStringParamList(r, ParamStreamID), userID)
	if err != nil {
		return RequestModifiers{}, err
	}
	result.ExcludeTargets, err = getStreams(request.QueryStringParamList(r, ParamStreamExcludes), userID)
	if err != nil {
		return RequestModifiers{}, err
	}

	result.FilterTargets, err = getStreams(request.QueryStringParamList(r, ParamStreamFilters), userID)
	if err != nil {
		return RequestModifiers{}, err
	}

	result.Count = request.QueryIntParam(r, ParamStreamMaxItems, 0)
	result.Offset = request.QueryIntParam(r, ParamContinuation, 0)
	result.StartTime = request.QueryInt64Param(r, ParamStreamStartTime, int64(0))
	result.StopTime = request.QueryInt64Param(r, ParamStreamStopTime, int64(0))
	return result, nil
}
