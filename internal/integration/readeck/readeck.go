// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readeck // import "miniflux.app/v2/internal/integration/readeck"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL string
	apiKey  string
	labels  string
	onlyURL bool
}

func NewClient(baseURL, apiKey, labels string, onlyURL bool) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, labels: labels, onlyURL: onlyURL}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string, entryContent string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("readeck: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return fmt.Errorf(`readeck: invalid API endpoint: %v`, err)
	}

	labelsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}
	labelsSplit := strings.FieldsFunc(c.labels, labelsSplitFn)

	var request *http.Request
	if c.onlyURL {
		requestBodyJson, err := json.Marshal(&readeckBookmark{
			Url:    entryURL,
			Title:  entryTitle,
			Labels: labelsSplit,
		})
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body: %v", err)
		}
		request, err = http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBodyJson))
		if err != nil {
			return fmt.Errorf("readeck: unable to create request: %v", err)
		}
		request.Header.Set("Content-Type", "application/json")
	} else {
		requestBody := new(bytes.Buffer)
		multipartWriter := multipart.NewWriter(requestBody)

		urlPart, err := multipartWriter.CreateFormField("url")
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body (entry url): %v", err)
		}
		urlPart.Write([]byte(entryURL))

		titlePart, err := multipartWriter.CreateFormField("title")
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body (entry title): %v", err)
		}
		titlePart.Write([]byte(entryTitle))

		featurePart, err := multipartWriter.CreateFormField("feature_find_main")
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body (feature_find_main flag): %v", err)
		}
		featurePart.Write([]byte("false")) // false to disable readability

		for _, label := range labelsSplit {
			labelPart, err := multipartWriter.CreateFormField("labels")
			if err != nil {
				return fmt.Errorf("readeck: unable to encode request body (entry labels): %v", err)
			}
			labelPart.Write([]byte(label))
		}

		contentBodyHeader, err := json.Marshal(&partContentHeader{
			Url:           entryURL,
			ContentHeader: contentHeader{ContentType: "text/html"},
		})
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body (entry content header): %v", err)
		}

		contentPart, err := multipartWriter.CreateFormFile("resource", "blob")
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body (entry content): %v", err)
		}
		contentPart.Write(contentBodyHeader)
		contentPart.Write([]byte("\n"))
		contentPart.Write([]byte(entryContent))

		err = multipartWriter.Close()
		if err != nil {
			return fmt.Errorf("readeck: unable to encode request body: %v", err)
		}
		request, err = http.NewRequest(http.MethodPost, apiEndpoint, requestBody)
		if err != nil {
			return fmt.Errorf("readeck: unable to create request: %v", err)
		}
		request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	}

	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("readeck: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("readeck: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type readeckBookmark struct {
	Url    string   `json:"url"`
	Title  string   `json:"title"`
	Labels []string `json:"labels,omitempty"`
}

type contentHeader struct {
	ContentType string `json:"content-type"`
}

type partContentHeader struct {
	Url           string        `json:"url"`
	ContentHeader contentHeader `json:"headers"`
}
