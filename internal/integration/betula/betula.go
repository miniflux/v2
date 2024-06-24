package betula

import (
	"fmt"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
	"net/http"
	"strings"
	"time"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	url      string
	username string
	password string
	token    string
}

func NewClient(url, username, password, token string) *Client {
	return &Client{url: url, username: username, password: password, token: token}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string, tags []string) error {
	var err error
	if c.token == "" {
		c.token, err = c.getToken()
		if err != nil {
			return err
		}
	}
	return c.createEntry(entryURL, entryTitle, tags)
}

func (c *Client) createEntry(entryURL, entryTitle string, tags []string) error {
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.url, "/save-link")
	if err != nil {
		return fmt.Errorf("betula: unable to generate save-link endpoint: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, nil)
	if err != nil {
		return fmt.Errorf("betula: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.AddCookie(&http.Cookie{Name: "betula-token", Value: c.token})

	request.Form.Add("url", entryURL)
	request.Form.Add("title", entryTitle)
	request.Form.Add("tags", strings.Join(tags, ","))

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

func (c *Client) getToken() (string, error) {
	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.url, "/login")
	if err != nil {
		return "", fmt.Errorf("betula: unable to generate token endpoint: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("betula: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	request.Form.Add("name", c.username)
	request.Form.Add("pass", c.password)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("betula: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("betula: unable to get token: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	if response.Cookies() == nil {
		return "", fmt.Errorf("betula: missing cookies")
	}

	return getCookie(response.Cookies(), "betula-token"), nil
}

func getCookie(cookies []*http.Cookie, key string) string {
	for _, c := range cookies {
		if c.Name == key {
			return c.Value
		}
	}
	return ""
}
