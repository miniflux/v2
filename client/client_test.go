// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newFakeHTTPClient(
	t *testing.T,
	fn func(t *testing.T, req *http.Request) *http.Response,
) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				return fn(t, req), nil
			}),
	}
}

func jsonResponseFrom(
	t *testing.T,
	status int,
	headers http.Header,
	body any,
) *http.Response {
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Unable to marshal body: %v", err)
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBuffer(data)),
		Header:     headers,
	}
}

func asJSON(data any) string {
	json, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(json)
}

func expectRequest(
	t *testing.T,
	method string,
	url string,
	checkBody func(r io.Reader),
	req *http.Request,
) {
	if req.Method != method {
		t.Fatalf("Expected method to be %s, got %s", method, req.Method)
	}

	if req.URL.String() != url {
		t.Fatalf("Expected URL path to be %s, got %s", url, req.URL)
	}

	if checkBody != nil {
		checkBody(req.Body)
	}
}

func expectFromJSON[T any](
	t *testing.T,
	r io.Reader,
	expected *T,
) {
	var got T
	if err := json.NewDecoder(r).Decode(&got); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(&got, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(got))
	}
}

func TestHealthcheck(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/healthcheck", nil, req)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
				}
			})))
	if err := client.HealthcheckContext(t.Context()); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestVersion(t *testing.T) {
	expected := &VersionResponse{
		Version:   "1.0.0",
		Commit:    "1234567890",
		BuildDate: "2021-01-01T00:00:00Z",
		GoVersion: "go1.20",
		Compiler:  "gc",
		Arch:      "amd64",
		OS:        "linux",
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/version", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.VersionContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestMe(t *testing.T) {
	expected := &User{
		ID:                        1,
		Username:                  "test",
		Password:                  "password",
		IsAdmin:                   false,
		Theme:                     "light",
		Language:                  "en",
		Timezone:                  "UTC",
		EntryDirection:            "asc",
		EntryOrder:                "created_at",
		Stylesheet:                "default",
		CustomJS:                  "custom.js",
		GoogleID:                  "google-id",
		OpenIDConnectID:           "openid-connect-id",
		EntriesPerPage:            10,
		KeyboardShortcuts:         true,
		ShowReadingTime:           true,
		EntrySwipe:                true,
		GestureNav:                "horizontal",
		DisplayMode:               "read",
		DefaultReadingSpeed:       1,
		CJKReadingSpeed:           1,
		DefaultHomePage:           "home",
		CategoriesSortingOrder:    "asc",
		MarkReadOnView:            true,
		MediaPlaybackRate:         1.0,
		BlockFilterEntryRules:     "block",
		KeepFilterEntryRules:      "keep",
		ExternalFontHosts:         "https://fonts.googleapis.com",
		AlwaysOpenExternalLinks:   true,
		OpenExternalLinksInNewTab: true,
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/me", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.MeContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUsers(t *testing.T) {
	expected := Users{
		{
			ID:       1,
			Username: "test1",
		},
		{
			ID:       2,
			Username: "test2",
		},
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/users", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UsersContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUserByID(t *testing.T) {
	expected := &User{
		ID:       1,
		Username: "test",
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/users/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UserByIDContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUserByUsername(t *testing.T) {
	expected := &User{
		ID:       1,
		Username: "test",
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/users/test", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UserByUsernameContext(t.Context(), "test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCreateUser(t *testing.T) {
	expected := &User{
		ID:       1,
		Username: "test",
		Password: "password",
		IsAdmin:  true,
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				exp := UserCreationRequest{
					Username: "test",
					Password: "password",
					IsAdmin:  true,
				}
				expectRequest(
					t,
					http.MethodPost,
					"http://mf/v1/users",
					func(r io.Reader) {
						expectFromJSON(t, r, &exp)
					},
					req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CreateUserContext(t.Context(), "test", "password", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUpdateUser(t *testing.T) {
	expected := &User{
		ID:       1,
		Username: "test",
		Password: "password",
		IsAdmin:  true,
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/users/1", func(r io.Reader) {
					expectFromJSON(t, r, &UserModificationRequest{
						Username: &expected.Username,
						Password: &expected.Password,
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UpdateUserContext(t.Context(), 1, &UserModificationRequest{
		Username: &expected.Username,
		Password: &expected.Password,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestDeleteUser(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodDelete, "http://mf/v1/users/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.DeleteUserContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAPIKeys(t *testing.T) {
	expected := APIKeys{
		{
			ID:          1,
			Token:       "token",
			Description: "test",
		},
		{
			ID:          2,
			Token:       "token2",
			Description: "test2",
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/api-keys", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.APIKeysContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCreateAPIKey(t *testing.T) {
	expected := &APIKey{
		ID:          42,
		Token:       "some-token",
		Description: "desc",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/api-keys", func(r io.Reader) {
					expectFromJSON(t, r, &APIKeyCreationRequest{
						Description: "desc",
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CreateAPIKeyContext(t.Context(), "desc")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestDeleteAPIKey(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodDelete, "http://mf/v1/api-keys/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.DeleteAPIKeyContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkAllAsRead(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/users/1/mark-all-as-read", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.MarkAllAsReadContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestIntegrationsStatus(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/integrations/status", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, struct {
					HasIntegrations bool `json:"has_integrations"`
				}{
					HasIntegrations: true,
				})
			})))
	status, err := client.IntegrationsStatusContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !status {
		t.Fatalf("Expected integrations status to be true, got false")
	}
}

func TestDiscover(t *testing.T) {
	expected := Subscriptions{
		{
			URL:   "http://example.com",
			Title: "Example",
			Type:  "rss",
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/discover", func(r io.Reader) {
					expectFromJSON(t, r, &map[string]string{"url": "http://example.com"})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.DiscoverContext(t.Context(), "http://example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCategories(t *testing.T) {
	expected := Categories{
		{
			ID:    1,
			Title: "Example",
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/categories", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CategoriesContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCategoriesWithCounters(t *testing.T) {
	feedCount := 1
	totalUnread := 2
	expected := Categories{
		{
			ID:          1,
			Title:       "Example",
			FeedCount:   &feedCount,
			TotalUnread: &totalUnread,
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/categories?counts=true", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CategoriesWithCountersContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCreateCategory(t *testing.T) {
	expected := &Category{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/categories", func(r io.Reader) {
					expectFromJSON(t, r, &CategoryCreationRequest{
						Title: "Example",
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CreateCategoryContext(t.Context(), "Example")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestCreateCategoryWithOptions(t *testing.T) {
	expected := &Category{
		ID:           1,
		Title:        "Example",
		HideGlobally: true,
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/categories", func(r io.Reader) {
					expectFromJSON(t, r, &CategoryCreationRequest{
						Title:        "Example",
						HideGlobally: true,
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CreateCategoryWithOptionsContext(t.Context(), &CategoryCreationRequest{
		Title:        "Example",
		HideGlobally: true,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUpdateCategory(t *testing.T) {
	expected := &Category{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/categories/1", func(r io.Reader) {
					expectFromJSON(t, r, &CategoryModificationRequest{
						Title: &expected.Title,
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UpdateCategoryContext(t.Context(), 1, "Example")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestUpdateCategoryWithOptions(t *testing.T) {
	expected := &Category{
		ID:           1,
		Title:        "Example",
		HideGlobally: true,
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/categories/1", func(r io.Reader) {
					expectFromJSON(t, r, &CategoryModificationRequest{
						Title:        &expected.Title,
						HideGlobally: &expected.HideGlobally,
					})
				}, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UpdateCategoryWithOptionsContext(t.Context(), 1, &CategoryModificationRequest{
		Title:        &expected.Title,
		HideGlobally: &expected.HideGlobally,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestMarkCategoryAsRead(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/categories/1/mark-all-as-read", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.MarkCategoryAsReadContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCategoryFeeds(t *testing.T) {
	expected := Feeds{
		{
			ID:    1,
			Title: "Example",
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/categories/1/feeds", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CategoryFeedsContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestDeleteCategory(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodDelete, "http://mf/v1/categories/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.DeleteCategoryContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestRefreshCategory(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/categories/1/refresh", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.RefreshCategoryContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFeeds(t *testing.T) {
	expected := Feeds{
		{
			ID:                          1,
			Title:                       "Example",
			FeedURL:                     "http://example.com",
			SiteURL:                     "http://example.com",
			CheckedAt:                   time.Date(1970, 1, 1, 0, 7, 0, 0, time.UTC),
			Disabled:                    false,
			IgnoreHTTPCache:             false,
			AllowSelfSignedCertificates: false,
			FetchViaProxy:               false,
			ScraperRules:                "",
			RewriteRules:                "",
			UrlRewriteRules:             "",
			BlocklistRules:              "",
			KeeplistRules:               "",
			BlockFilterEntryRules:       "",
			KeepFilterEntryRules:        "",
			Crawler:                     false,
			UserAgent:                   "",
			Cookie:                      "",
			Username:                    "",
			Password:                    "",
			Category: &Category{
				ID:    1,
				Title: "Example",
			},
			HideGlobally: false,
			DisableHTTP2: false,
			ProxyURL:     "",
		},
		{
			ID:    2,
			Title: "Example 2",
		},
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FeedsContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestExport(t *testing.T) {
	expected := []byte("hello")
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/export", nil, req)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(string(expected))),
					Header:     http.Header{},
				}
			})))
	res, err := client.ExportContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, res)
	}
}

func TestImport(t *testing.T) {
	expected := []byte("hello")
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(
					t,
					http.MethodPost,
					"http://mf/v1/import",
					func(r io.Reader) {
						b, err := io.ReadAll(r)
						if err != nil {
							t.Fatalf("Expected no error, got %v", err)
						}
						if !bytes.Equal(b, expected) {
							t.Fatalf("expected %+v, got %+v", expected, b)
						}
					},
					req)
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{},
				}
			})))
	if err := client.ImportContext(t.Context(), io.NopCloser(bytes.NewBufferString(string(expected)))); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFeed(t *testing.T) {
	expected := &Feed{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FeedContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestCreateFeed(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/feeds", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, struct {
					FeedID int64 `json:"feed_id"`
				}{
					FeedID: 1,
				})
			})))
	id, err := client.CreateFeedContext(t.Context(), &FeedCreationRequest{
		FeedURL: "http://example.com",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id != 1 {
		t.Fatalf("Expected feed ID to be 1, got %d", id)
	}
}

func TestUpdateFeed(t *testing.T) {
	expected := &Feed{
		ID:      1,
		FeedURL: "http://example.com/",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/feeds/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UpdateFeedContext(t.Context(), 1, &FeedModificationRequest{
		FeedURL: &expected.FeedURL,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestMarkFeedAsRead(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/feeds/1/mark-all-as-read", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.MarkFeedAsReadContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestRefreshAllFeeds(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/feeds/refresh", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.RefreshAllFeedsContext(t.Context()); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestRefreshFeed(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/feeds/1/refresh", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.RefreshFeedContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestForceRefreshFeed(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/feeds/1/refresh?force=true", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.ForceRefreshFeedContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteFeed(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodDelete, "http://mf/v1/feeds/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.DeleteFeedContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFeedIcon(t *testing.T) {
	expected := &FeedIcon{
		ID:       1,
		MimeType: "text/plain",
		Data:     "data",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds/1/icon", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FeedIconContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestFeedEntry(t *testing.T) {
	expected := &Entry{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds/1/entries/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FeedEntryContext(t.Context(), 1, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestCategoryEntry(t *testing.T) {
	expected := &Entry{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/categories/1/entries/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CategoryEntryContext(t.Context(), 1, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntry(t *testing.T) {
	expected := &Entry{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntries(t *testing.T) {
	expected := &EntryResultSet{
		Total: 1,
		Entries: Entries{
			{
				ID:    1,
				Title: "Example",
			},
		},
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntriesContext(t.Context(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestFeedEntries(t *testing.T) {
	expected := &EntryResultSet{
		Total: 1,
		Entries: Entries{
			{
				ID:    1,
				Title: "Example",
			},
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds/1/entries?limit=10&offset=0", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FeedEntriesContext(t.Context(), 1, &Filter{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestCategoryEntries(t *testing.T) {
	expected := &EntryResultSet{
		Total: 1,
		Entries: Entries{
			{
				ID:    1,
				Title: "Example",
			},
		},
	}

	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/categories/1/entries?limit=10&offset=0", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.CategoryEntriesContext(t.Context(), 1, &Filter{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestUpdateEntries(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/entries", nil, req)
				expectFromJSON(t, req.Body, &struct {
					EntryIDs []int64 `json:"entry_ids"`
					Status   string  `json:"status"`
				}{
					EntryIDs: []int64{1, 2},
					Status:   "read",
				})
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.UpdateEntriesContext(t.Context(), []int64{1, 2}, "read"); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdateEntriesStarred(t *testing.T) {
	starred := true
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/entries", nil, req)
				expectFromJSON(t, req.Body, &struct {
					EntryIDs []int64 `json:"entry_ids"`
					Starred  *bool   `json:"starred"`
				}{
					EntryIDs: []int64{1, 2},
					Starred:  &starred,
				})
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.UpdateEntriesStarredContext(t.Context(), []int64{1, 2}, true); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdateEntry(t *testing.T) {
	expected := &Entry{
		ID:    1,
		Title: "Example",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/entries/1", nil, req)
				expectFromJSON(t, req.Body, &EntryModificationRequest{
					Title: &expected.Title,
				})
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.UpdateEntryContext(t.Context(), 1, &EntryModificationRequest{
		Title: &expected.Title,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestToggleStarred(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/entries/1/star", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.ToggleStarredContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSaveEntry(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPost, "http://mf/v1/entries/1/save", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.SaveEntryContext(t.Context(), 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFetchEntryOriginalContent(t *testing.T) {
	expected := "Example"
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/1/fetch-content", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, struct {
					Content string `json:"content"`
				}{
					Content: expected,
				})
			})))
	res, err := client.FetchEntryOriginalContentContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if res != expected {
		t.Fatalf("Expected %s, got %s", expected, res)
	}
}

func TestFetchCounters(t *testing.T) {
	expected := &FeedCounters{
		ReadCounters: map[int64]int{
			2: 1,
		},
		UnreadCounters: map[int64]int{
			3: 1,
		},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/feeds/counters", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.FetchCountersContext(t.Context())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestFlushHistory(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/flush-history", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, nil)
			})))
	if err := client.FlushHistoryContext(t.Context()); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestIcon(t *testing.T) {
	expected := &FeedIcon{
		ID:       1,
		MimeType: "text/plain",
		Data:     "data",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/icons/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.IconContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEnclosure(t *testing.T) {
	expected := &Enclosure{
		ID:       1,
		URL:      "http://example.com",
		MimeType: "text/plain",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/enclosures/1", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EnclosureContext(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestUpdateEnclosure(t *testing.T) {
	expected := &Enclosure{
		ID:       1,
		URL:      "http://example.com",
		MimeType: "text/plain",
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodPut, "http://mf/v1/enclosures/1", nil, req)
				expectFromJSON(t, req.Body, &EnclosureUpdateRequest{
					MediaProgression: 10,
				})
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	if err := client.UpdateEnclosureContext(t.Context(), 1, &EnclosureUpdateRequest{
		MediaProgression: 10,
	}); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func boolPtr(b bool) *bool { return &b }

func TestEntryIDsFilterSerialization(t *testing.T) {
	tests := []struct {
		name   string
		filter *EntryIDsFilter
		query  string
	}{
		{
			name:   "empty_filter",
			filter: &EntryIDsFilter{},
		},
		{
			name:   "non_positive_pagination",
			filter: &EntryIDsFilter{Limit: -1, Offset: -1},
		},
		{
			name:   "starred_false",
			filter: &EntryIDsFilter{Starred: boolPtr(false)},
			query:  "?starred=false",
		},
		{
			name:   "order",
			filter: &EntryIDsFilter{Order: "published_at"},
			query:  "?order=published_at",
		},
		{
			name:   "direction",
			filter: &EntryIDsFilter{Direction: "asc"},
			query:  "?direction=asc",
		},
		{
			name:   "feed_id",
			filter: &EntryIDsFilter{FeedID: 4294967297},
			query:  "?feed_id=4294967297",
		},
		{
			name:   "category_id",
			filter: &EntryIDsFilter{CategoryID: 4294967298},
			query:  "?category_id=4294967298",
		},
		{
			name:   "before",
			filter: &EntryIDsFilter{Before: 2147483648},
			query:  "?before=2147483648",
		},
		{
			name:   "after",
			filter: &EntryIDsFilter{After: 2147483649},
			query:  "?after=2147483649",
		},
		{
			name:   "published_before",
			filter: &EntryIDsFilter{PublishedBefore: 2147483650},
			query:  "?published_before=2147483650",
		},
		{
			name:   "published_after",
			filter: &EntryIDsFilter{PublishedAfter: 2147483651},
			query:  "?published_after=2147483651",
		},
		{
			name:   "changed_before",
			filter: &EntryIDsFilter{ChangedBefore: 2147483652},
			query:  "?changed_before=2147483652",
		},
		{
			name:   "changed_after",
			filter: &EntryIDsFilter{ChangedAfter: 2147483653},
			query:  "?changed_after=2147483653",
		},
		{
			name:   "before_entry_id",
			filter: &EntryIDsFilter{BeforeEntryID: 4294967299},
			query:  "?before_entry_id=4294967299",
		},
		{
			name:   "after_entry_id",
			filter: &EntryIDsFilter{AfterEntryID: 4294967300},
			query:  "?after_entry_id=4294967300",
		},
		{
			name:   "search_encoding",
			filter: &EntryIDsFilter{Search: `go client & "feeds" + tags`},
			query:  "?search=go+client+%26+%22feeds%22+%2B+tags",
		},
		{
			name:   "repeated_statuses",
			filter: &EntryIDsFilter{Statuses: []string{EntryStatusRead, EntryStatusUnread, EntryStatusRead}},
			query:  "?status=read&status=unread&status=read",
		},
		{
			name:   "statuses_override_status",
			filter: &EntryIDsFilter{Status: "removed", Statuses: []string{EntryStatusRead, EntryStatusUnread}},
			query:  "?status=read&status=unread",
		},
		{
			name:   "empty_statuses_preserve_status",
			filter: &EntryIDsFilter{Status: EntryStatusUnread, Statuses: []string{}},
			query:  "?status=unread",
		},
		{
			name:   "repeated_tags",
			filter: &EntryIDsFilter{Tags: []string{"go", "release notes", "go"}},
			query:  "?tags=go&tags=release+notes&tags=go",
		},
		{
			name:   "globally_visible_true",
			filter: &EntryIDsFilter{GloballyVisible: boolPtr(true)},
			query:  "?globally_visible=true",
		},
		{
			name:   "globally_visible_false",
			filter: &EntryIDsFilter{GloballyVisible: boolPtr(false)},
			query:  "?globally_visible=false",
		},
		{
			name:   "zero_boundaries",
			filter: &EntryIDsFilter{Status: EntryStatusUnread},
			query:  "?status=unread",
		},
		{
			name:   "negative_feed_id",
			filter: &EntryIDsFilter{FeedID: -1},
			query:  "?feed_id=-1",
		},
		{
			name:   "negative_category_id",
			filter: &EntryIDsFilter{CategoryID: -2},
			query:  "?category_id=-2",
		},
		{
			name:   "negative_before",
			filter: &EntryIDsFilter{Before: -3},
			query:  "?before=-3",
		},
		{
			name:   "negative_after",
			filter: &EntryIDsFilter{After: -4},
			query:  "?after=-4",
		},
		{
			name:   "negative_published_before",
			filter: &EntryIDsFilter{PublishedBefore: -5},
			query:  "?published_before=-5",
		},
		{
			name:   "negative_published_after",
			filter: &EntryIDsFilter{PublishedAfter: -6},
			query:  "?published_after=-6",
		},
		{
			name:   "negative_changed_before",
			filter: &EntryIDsFilter{ChangedBefore: -7},
			query:  "?changed_before=-7",
		},
		{
			name:   "negative_changed_after",
			filter: &EntryIDsFilter{ChangedAfter: -8},
			query:  "?changed_after=-8",
		},
		{
			name:   "negative_before_entry_id",
			filter: &EntryIDsFilter{BeforeEntryID: -9},
			query:  "?before_entry_id=-9",
		},
		{
			name:   "negative_after_entry_id",
			filter: &EntryIDsFilter{AfterEntryID: -10},
			query:  "?after_entry_id=-10",
		},
	}

	expected := &EntryIDsResultSet{Total: 1, EntryIDs: []int64{42}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithOptions(
				"http://mf",
				WithHTTPClient(
					newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
						expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids"+tt.query, nil, req)
						return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
					})))
			res, err := client.EntryIDsContext(t.Context(), tt.filter)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if !reflect.DeepEqual(res, expected) {
				t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
			}
		})
	}
}

func TestEntryIDsContextBadRequest(t *testing.T) {
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids?feed_id=-1&status=unread", nil, req)
				return jsonResponseFrom(t, http.StatusBadRequest, http.Header{}, errorResponse{
					ErrorMessage: "Invalid feed_id query parameter",
				})
			})))
	res, err := client.EntryIDsContext(t.Context(), &EntryIDsFilter{Status: EntryStatusUnread, FeedID: -1})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("Expected ErrBadRequest, got %v", err)
	}
	if res != nil {
		t.Fatalf("Expected no result on error, got %s", asJSON(res))
	}
}

func TestEntryIDsWithAllFilters(t *testing.T) {
	filter := &EntryIDsFilter{
		Limit:           2,
		Offset:          5,
		Starred:         boolPtr(false),
		Status:          "removed",
		Order:           "published_at",
		Direction:       "asc",
		FeedID:          42,
		CategoryID:      7,
		Before:          1800000000,
		After:           1700000000,
		PublishedBefore: 1790000000,
		PublishedAfter:  1710000000,
		ChangedBefore:   1780000000,
		ChangedAfter:    1720000000,
		BeforeEntryID:   100,
		AfterEntryID:    10,
		Search:          `go & "rss"`,
		Statuses:        []string{EntryStatusUnread, EntryStatusRead},
		Tags:            []string{"go", "release notes"},
		GloballyVisible: boolPtr(false),
	}
	expected := &EntryIDsResultSet{Total: 1, EntryIDs: []int64{42}}
	const expectedURL = "http://mf/v1/entries/ids?after=1700000000&after_entry_id=10" +
		"&before=1800000000&before_entry_id=100&category_id=7" +
		"&changed_after=1720000000&changed_before=1780000000&direction=asc" +
		"&feed_id=42&globally_visible=false&limit=2&offset=5&order=published_at" +
		"&published_after=1710000000&published_before=1790000000" +
		"&search=go+%26+%22rss%22&starred=false&status=unread&status=read" +
		"&tags=go&tags=release+notes"
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, expectedURL, nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDs(filter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntryIDsNoFilter(t *testing.T) {
	expected := &EntryIDsResultSet{
		Total:    2,
		EntryIDs: []int64{1, 2},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDsContext(t.Context(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntryIDsWithPaginationFilter(t *testing.T) {
	expected := &EntryIDsResultSet{
		Total:    5,
		EntryIDs: []int64{3},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids?limit=1&offset=2", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDsContext(t.Context(), &EntryIDsFilter{Limit: 1, Offset: 2})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntryIDsWithStarredFilter(t *testing.T) {
	expected := &EntryIDsResultSet{
		Total:    1,
		EntryIDs: []int64{42},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids?starred=true", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDsContext(t.Context(), &EntryIDsFilter{Starred: boolPtr(true)})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntryIDsWithStatusFilter(t *testing.T) {
	expected := &EntryIDsResultSet{
		Total:    10,
		EntryIDs: []int64{7, 8},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids?status=unread", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDsContext(t.Context(), &EntryIDsFilter{Status: EntryStatusUnread})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}

func TestEntryIDsWithCombinedFilter(t *testing.T) {
	expected := &EntryIDsResultSet{
		Total:    3,
		EntryIDs: []int64{5},
	}
	client := NewClientWithOptions(
		"http://mf",
		WithHTTPClient(
			newFakeHTTPClient(t, func(t *testing.T, req *http.Request) *http.Response {
				expectRequest(t, http.MethodGet, "http://mf/v1/entries/ids?limit=2&offset=5&starred=false&status=read", nil, req)
				return jsonResponseFrom(t, http.StatusOK, http.Header{}, expected)
			})))
	res, err := client.EntryIDsContext(t.Context(), &EntryIDsFilter{Limit: 2, Offset: 5, Starred: boolPtr(false), Status: EntryStatusRead})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected %s, got %s", asJSON(expected), asJSON(res))
	}
}
