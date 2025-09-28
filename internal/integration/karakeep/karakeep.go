// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package karakeep // import "miniflux.app/v2/internal/integration/karakeep"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	wrapped     *http.Client
	apiEndpoint string
	apiToken    string
	tags        string
}

type tagItem struct {
	TagName string `json:"tagName"`
}

type saveURLPayload struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type saveURLResponse struct {
	ID string `json:"id"`
}

type attachTagsPayload struct {
	Tags []tagItem `json:"tags"`
}

type errorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func NewClient(apiToken string, apiEndpoint string, tags string) *Client {
	return &Client{wrapped: &http.Client{Timeout: defaultClientTimeout}, apiEndpoint: apiEndpoint, apiToken: apiToken, tags: tags}
}

func (c *Client) attachTags(entryID string) error {
	if c.tags == "" {
		return nil
	}

	tagItems := make([]tagItem, 0)
	for tag := range strings.SplitSeq(c.tags, ",") {
		if trimmedTag := strings.TrimSpace(tag); trimmedTag != "" {
			tagItems = append(tagItems, tagItem{TagName: trimmedTag})
		}
	}

	if len(tagItems) == 0 {
		return nil
	}

	tagRequestBody, err := json.Marshal(&attachTagsPayload{
		Tags: tagItems,
	})
	if err != nil {
		return fmt.Errorf("karakeep: unable to encode tag request body: %v", err)
	}

	tagRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/tags", c.apiEndpoint, entryID), bytes.NewReader(tagRequestBody))
	if err != nil {
		return fmt.Errorf("karakeep: unable to create tag request: %v", err)
	}

	tagRequest.Header.Set("Authorization", "Bearer "+c.apiToken)
	tagRequest.Header.Set("Content-Type", "application/json")
	tagRequest.Header.Set("User-Agent", "Miniflux/"+version.Version)

	tagResponse, err := c.wrapped.Do(tagRequest)
	if err != nil {
		return fmt.Errorf("karakeep: unable to send tag request: %v", err)
	}
	defer tagResponse.Body.Close()

	if tagResponse.StatusCode != http.StatusOK && tagResponse.StatusCode != http.StatusCreated {
		tagResponseBody, err := io.ReadAll(tagResponse.Body)
		if err != nil {
			return fmt.Errorf("karakeep: failed to parse tag response: %s", err)
		}

		var errResponse errorResponse
		if err := json.Unmarshal(tagResponseBody, &errResponse); err != nil {
			return fmt.Errorf("karakeep: unable to parse tag error response: status=%d body=%s", tagResponse.StatusCode, string(tagResponseBody))
		}
		return fmt.Errorf("karakeep: failed to attach tags: status=%d errorcode=%s %s", tagResponse.StatusCode, errResponse.Code, errResponse.Error)
	}

	return nil
}

func (c *Client) SaveURL(entryURL string) error {
	requestBody, err := json.Marshal(&saveURLPayload{
		Type: "link",
		URL:  entryURL,
	})
	if err != nil {
		return fmt.Errorf("karakeep: unable to encode request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("karakeep: unable to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	resp, err := c.wrapped.Do(req)
	if err != nil {
		return fmt.Errorf("karakeep: unable to send request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("karakeep: failed to parse response: %s", err)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("karakeep: unexpected content type response: %s", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != http.StatusCreated {
		var errResponse errorResponse
		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			return fmt.Errorf("karakeep: unable to parse error response: status=%d body=%s", resp.StatusCode, string(responseBody))
		}
		return fmt.Errorf("karakeep: failed to save URL: status=%d errorcode=%s %s", resp.StatusCode, errResponse.Code, errResponse.Error)
	}

	var response saveURLResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("karakeep: unable to parse response: %v", err)
	}

	if response.ID == "" {
		return errors.New("karakeep: unable to get ID from response")
	}

	if err := c.attachTags(response.ID); err != nil {
		return fmt.Errorf("karakeep: unable to attach tags: %v", err)
	}

	return nil
}
