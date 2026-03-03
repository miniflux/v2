// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ai // import "miniflux.app/v2/internal/integration/ai"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	defaultClientTimeout = 30 * time.Second
	maxContentLength     = 4000

	// systemPrompt instructs the AI to return a JSON object with summary and score.
	// Do NOT make this user-configurable — it is intentionally hardcoded.
	systemPrompt = "You are a content analyzer. For the given article, provide:\n" +
		"1. A concise summary in 2-3 sentences in the same language as the article\n" +
		"2. A relevance/quality score from 1 to 10 (10=must-read, 1=skip)\n" +
		"Respond ONLY with JSON: {\"summary\": \"...\", \"score\": N}"
)

var htmlTagRegexp = regexp.MustCompile("<[^>]*>")

// Client communicates with an OpenAI-compatible chat completions API.
type Client struct {
	providerURL string // e.g. "https://api.openai.com/v1"
	apiKey      string
	model       string
}

// NewClient creates a new AI client for the given OpenAI-compatible provider.
func NewClient(providerURL, apiKey, model string) *Client {
	return &Client{
		providerURL: providerURL,
		apiKey:      apiKey,
		model:       model,
	}
}

// SummarizeResult holds the AI-generated summary and score for an article.
type SummarizeResult struct {
	Summary string `json:"summary"`
	Score   int    `json:"score"`
}

// SummarizeEntry sends article content to the AI provider and returns a summary with a 1-10 score.
// It calls the OpenAI-compatible /chat/completions endpoint.
// The content is truncated to ~4000 chars to control token usage.
// If the entry already has a summary (non-empty aiSummary), it returns nil to avoid wasting tokens.
func (c *Client) SummarizeEntry(title, content, aiSummary string) (*SummarizeResult, error) {
	// Skip if already summarized — avoid duplicate API calls and wasted tokens.
	if aiSummary != "" {
		return nil, nil
	}

	cleanContent := truncateContent(stripHTMLTags(content), maxContentLength)
	userMessage := title + "\n\n" + cleanContent

	requestPayload := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.3,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("ai: unable to encode request body: %v", err)
	}

	apiEndpoint := strings.TrimRight(c.providerURL, "/") + "/chat/completions"
	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("ai: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Use system proxy settings (HTTP_PROXY, HTTPS_PROXY, NO_PROXY env vars).
	httpClient := &http.Client{
		Timeout: defaultClientTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("ai: unable to send request: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ai: unable to read response body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai: provider returned status %d: %s", response.StatusCode, truncateContent(string(responseBody), 512))
	}

	var completionResp chatCompletionResponse
	if err := json.Unmarshal(responseBody, &completionResp); err != nil {
		return nil, fmt.Errorf("ai: unable to parse response JSON: %v", err)
	}

	if len(completionResp.Choices) == 0 {
		return nil, fmt.Errorf("ai: response contains no choices")
	}

	messageContent := strings.TrimSpace(completionResp.Choices[0].Message.Content)
	if messageContent == "" {
		return nil, fmt.Errorf("ai: response message content is empty")
	}

	// The AI message content is itself a JSON string — parse it into SummarizeResult.
	var result SummarizeResult
	if err := json.Unmarshal([]byte(messageContent), &result); err != nil {
		return nil, fmt.Errorf("ai: unable to parse summary JSON from model response: %v (raw: %s)", err, truncateContent(messageContent, 256))
	}

	// Clamp score to valid 1-10 range to handle model hallucinations.
	if result.Score < 1 {
		result.Score = 1
	}
	if result.Score > 10 {
		result.Score = 10
	}

	return &result, nil
}

// stripHTMLTags removes HTML tags from content for AI consumption.
// This is a simple approach — not a full sanitizer, just for truncation purposes.
func stripHTMLTags(s string) string {
	cleaned := htmlTagRegexp.ReplaceAllString(s, " ")
	// Collapse multiple whitespace into single space.
	spaceRegexp := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(spaceRegexp.ReplaceAllString(cleaned, " "))
}

// truncateContent limits content to maxLen characters to control token usage.
func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// chatMessage represents a single message in the chat completions request.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionRequest is the request body for the OpenAI-compatible chat completions endpoint.
type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// chatCompletionResponse is the response from the OpenAI-compatible chat completions endpoint.
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
