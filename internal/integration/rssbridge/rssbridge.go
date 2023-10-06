package rssbridge // import "miniflux.app/integration/rssbridge"

import (
	"net/url"
	"regexp"

	"miniflux.app/v2/internal/errors"
	"miniflux.app/v2/internal/http/client"
)

func DetectBridge(rssbridgeURL, websiteURL string) (string, *errors.LocalizedError) {
	u, err := url.Parse(rssbridgeURL)
	if err != nil {
		return "", errors.NewLocalizedError("rss-bridge: invalid url", err)
	}
	values := u.Query()
	values.Add("action", "detect")
	values.Add("format", "html")
	values.Add("url", websiteURL)
	u.RawQuery = values.Encode()

	clt := client.New(u.String())

	response, err := clt.Get()
	if err != nil {
		return "", errors.NewLocalizedError("RSS-Bridge: %v", err)
	}
	body := response.BodyAsString()
	if response.HasServerFailure() {
		r, _ := regexp.Compile("<strong>Message:</strong>(.+)</div>")
		if matches := r.FindStringSubmatch(body); len(matches) > 1 {
			return "", errors.NewLocalizedError("RSS-Bridge: %v", matches[1])
		}
		return "", errors.NewLocalizedError("RSS-Bridge: Server Failure")
	}
	return body, nil
}
