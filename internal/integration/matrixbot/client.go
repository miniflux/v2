// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package matrixbot // import "miniflux.app/v2/internal/integration/matrixbot"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	matrixBaseURL string
}

func NewClient(matrixBaseURL string) *Client {
	return &Client{matrixBaseURL: matrixBaseURL}
}

// Specs: https://spec.matrix.org/v1.8/client-server-api/#getwell-knownmatrixclient
func (c *Client) DiscoverEndpoints() (*DiscoveryEndpointResponse, error) {
	endpointURL, err := url.JoinPath(c.matrixBaseURL, "/.well-known/matrix/client")
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to join base URL and path: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, endpointURL, nil)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to create request: %v", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("matrix: unexpected response from %s status code is %d", endpointURL, response.StatusCode)
	}

	var discoveryEndpointResponse DiscoveryEndpointResponse
	if err := json.NewDecoder(response.Body).Decode(&discoveryEndpointResponse); err != nil {
		return nil, fmt.Errorf("matrix: unable to decode discovery response: %w", err)
	}

	return &discoveryEndpointResponse, nil
}

// Specs https://spec.matrix.org/v1.8/client-server-api/#post_matrixclientv3login
func (c *Client) Login(homeServerURL, matrixUsername, matrixPassword string) (*LoginResponse, error) {
	endpointURL, err := url.JoinPath(homeServerURL, "/_matrix/client/v3/login")
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to join base URL and path: %w", err)
	}

	loginRequest := LoginRequest{
		Type: "m.login.password",
		Identifier: UserIdentifier{
			Type: "m.id.user",
			User: matrixUsername,
		},
		Password: matrixPassword,
	}

	requestBody, err := json.Marshal(loginRequest)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("matrix: unexpected response from %s status code is %d", endpointURL, response.StatusCode)
	}

	var loginResponse LoginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		return nil, fmt.Errorf("matrix: unable to decode login response: %w", err)
	}

	return &loginResponse, nil
}

// Specs https://spec.matrix.org/v1.8/client-server-api/#put_matrixclientv3roomsroomidsendeventtypetxnid
func (c *Client) SendFormattedTextMessage(homeServerURL, accessToken, roomID, textMessage, formattedMessage string) (*RoomEventResponse, error) {
	txnID := crypto.GenerateRandomStringHex(10)
	endpointURL, err := url.JoinPath(homeServerURL, "/_matrix/client/v3/rooms/", roomID, "/send/m.room.message/", txnID)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to join base URL and path: %w", err)
	}

	messageEvent := TextMessageEventRequest{
		MsgType:       "m.text",
		Body:          textMessage,
		Format:        "org.matrix.custom.html",
		FormattedBody: formattedMessage,
	}

	requestBody, err := json.Marshal(messageEvent)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPut, endpointURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("matrix: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("matrix: unexpected response from %s status code is %d", endpointURL, response.StatusCode)
	}

	var eventResponse RoomEventResponse
	if err := json.NewDecoder(response.Body).Decode(&eventResponse); err != nil {
		return nil, fmt.Errorf("matrix: unable to decode event response: %w", err)
	}

	return &eventResponse, nil
}

type HomeServerInformation struct {
	BaseURL string `json:"base_url"`
}

type IdentityServerInformation struct {
	BaseURL string `json:"base_url"`
}

type DiscoveryEndpointResponse struct {
	HomeServerInformation     HomeServerInformation     `json:"m.homeserver"`
	IdentityServerInformation IdentityServerInformation `json:"m.identity_server"`
}

type UserIdentifier struct {
	Type string `json:"type"`
	User string `json:"user"`
}

type LoginRequest struct {
	Type       string         `json:"type"`
	Identifier UserIdentifier `json:"identifier"`
	Password   string         `json:"password"`
}

type LoginResponse struct {
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token"`
	DeviceID    string `json:"device_id"`
	HomeServer  string `json:"home_server"`
}

type TextMessageEventRequest struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

type RoomEventResponse struct {
	EventID string `json:"event_id"`
}
