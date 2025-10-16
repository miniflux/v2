// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
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
