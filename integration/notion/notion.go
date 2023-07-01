package notion

import (
	"fmt"
	"net/url"

	"miniflux.app/http/client"
)

// Client represents a Notion client.
type Client struct {
	baseURL string
	token   string
	pageID  string
}

// NewClient returns a new Notion client.
func NewClient(baseURL, token string, pageID string) *Client {
	return &Client{baseURL, token, pageID}
}

func (c *Client) AddEntry(entryURL string, entryTitle string) error {
	if c.baseURL == "" || c.token == "" || c.pageID == "" {
		return fmt.Errorf("notion: missing credentials")
	}
	endpoint, err := getAPIEndpoint(c.baseURL, "/v1/blocks/"+c.pageID+"/children")
	if err != nil {
		return fmt.Errorf("notion: unable to get token endpoint: %v", err)
	}

	clt := client.New(endpoint)
	block := &Data{
		Children: []Block{
			{
				Object: "block",
				Type:   "bookmark",
				Bookmark: Bookmark{
					Caption: []interface{}{},
					URL:     entryURL,
				},
			},
		},
	}
	clt.WithAuthorization("Bearer " + c.token)
	customHeaders := map[string]string{
		"Notion-Version": "2022-06-28",
	}
	clt.WithCustomHeaders(customHeaders)
	response, error := clt.PatchJSON(block)
	if error != nil {
		return fmt.Errorf("notion: unable to post entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("notion: request failed, status=%d", response.StatusCode)
	}
	return nil

}

func getAPIEndpoint(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("notion: invalid API endpoint: %v", err)
	}
	u.Path = path
	return u.String(), nil
}
