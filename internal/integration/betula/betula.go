package betula

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	url   string
	token string
}

func NewClient(url, token string) *Client {
	return &Client{url: url, token: token}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string, tags []string) error {
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.url, "/save-link")
	if err != nil {
		return fmt.Errorf("betula: unable to generate save-link endpoint: %v", err)
	}

	values := url.Values{}
	values.Add("url", entryURL)
	values.Add("title", entryTitle)
	values.Add("tags", strings.Join(tags, ","))

	request, err := http.NewRequest(http.MethodPost, apiEndpoint+"?"+values.Encode(), nil)
	if err != nil {
		return fmt.Errorf("betula: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.AddCookie(&http.Cookie{Name: "betula-token", Value: c.token})

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("betula: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("betula: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}
