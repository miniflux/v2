// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package siyuannote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	URL          string
	NotebookName string
	PagePath     string
	Token        string
}

type NotebookData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Notebooks []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Icon   string `json:"icon"`
			Sort   int    `json:"sort"`
			Closed bool   `json:"closed"`
		} `json:"notebooks"`
	} `json:"data"`
}

type DocumentGeneration struct {
	Notebook string `json:"notebook"`
	Path     string `json:"path"`
	Markdown string `json:"markdown"`
}

func NewClient(URL string, apiToken, notebookName string, pagePath string) *Client {
	return &Client{URL, notebookName, pagePath, apiToken}
}

func (c *Client) UpdateDocument(entryURL string, entryTitle string, entryDocument string) error {
	if c.URL == "" {
		return fmt.Errorf("SiyuanNote Integration: missing URL")
	}

	notebookApiEndpoint := c.URL + "/api/notebook/lsNotebooks"
	documentApiEndpoint := c.URL + "/api/filetree/createDocWithMd"

	// Get the list of notebooks to get the right ID
	request, error := http.Post(notebookApiEndpoint, "application/json", nil)
	request.Header.Set("Authorization", c.Token)
	if error != nil {
		return fmt.Errorf("SiyuanNote Integration: Error during HTTP post request to siyuan %s", error)
	}

	defer request.Body.Close()

	if request.StatusCode == http.StatusOK {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			//Failed to read response.
			return fmt.Errorf("SiyuanNote Integration: Error during HTTP post reading response %s", err)
		}

		var notebook_data NotebookData
		json_error := json.Unmarshal(body, &notebook_data)

		if json_error != nil {
			return fmt.Errorf("SiyuanNote Integration: Error during JSON Unmarshalling %s", json_error)
		}
		var notebook_found bool

		// Scan the list of notebooks to find the correct one
		for _, value := range notebook_data.Data.Notebooks {
			if value.Name == c.NotebookName {
				notebook_found = true

				// Create the RSS document under the requested path
				var document DocumentGeneration
				document.Notebook = value.ID

				document.Path = c.PagePath + "/" + entryTitle

				document.Markdown = "---\n<" + entryURL + ">\n" + "---\n" + entryDocument
				json_data, json_error := json.Marshal(document)
				if json_error != nil {
					return fmt.Errorf("SiyuanNote Integration: Error during JSON Marshalling %s", json_error)
				}

				res, error := http.Post(documentApiEndpoint, "application/json", bytes.NewBuffer(json_data))
				if error != nil {
					return fmt.Errorf("SiyuanNote Integration: Error during Document Creation Post Request %s", error)
				}
				defer res.Body.Close()
				if request.StatusCode == http.StatusOK {
					_, err := io.ReadAll(res.Body)
					if err != nil {
						//Failed to read response.
						return fmt.Errorf("SiyuanNote Integration: Error during Document Creation Post response reading %s", error)
					}
				} else {
					return fmt.Errorf("SiyuanNote Integration: Error during Document Creation Post request %d", request.StatusCode)
				}
			}
		}
		if notebook_found == false {
			return fmt.Errorf("SiyuanNote Integration: The notebook %s not found ", c.NotebookName)
		}
	} else {
		return fmt.Errorf("SiyuanNote Integration: Post Request to get list of notebooks failed with error code: %s", request.Status)
	}

	return nil
}
