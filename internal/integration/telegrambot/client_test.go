// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telegrambot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestErrorsDoNotContainBotToken(t *testing.T) {
	const secret = "123456:ABC-DEF1234ghIkl"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("unable to hijack connection: %v", err)
			return
		}
		connection.Close()
	}))
	defer server.Close()

	client := NewClient(secret, "-100123")
	client.apiEndpoint = server.URL

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "getMe",
			call: func() error {
				_, err := client.GetMe()
				return err
			},
		},
		{
			name: "sendMessage",
			call: func() error {
				_, err := client.SendMessage(&MessageRequest{ChatID: "-100123", Text: "Entry"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			if strings.Contains(err.Error(), secret) {
				t.Fatalf("expected the error to not contain the bot token, got %q", err)
			}

			if want := "telegram: unable to send request"; !strings.Contains(err.Error(), want) {
				t.Fatalf("expected the error to contain %q, got %q", want, err)
			}
		})
	}
}
