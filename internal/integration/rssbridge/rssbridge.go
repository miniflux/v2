package rssbridge // import "miniflux.app/integration/rssbridge"

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"

	"miniflux.app/v2/internal/errors"
)

func DetectBridge(rssbridgeURL, websiteURL string) (string, *errors.LocalizedError) {
	u, err := url.Parse(rssbridgeURL)
	if err != nil {
		return "", errors.NewLocalizedError("RSS-Bridge: invalid url", err)
	}
	values := u.Query()
	values.Add("action", "detect")
	values.Add("format", "html")
	values.Add("url", websiteURL)
	u.RawQuery = values.Encode()

	response, err := http.Get(u.String())
	if err != nil {
		return "", errors.NewLocalizedError("RSS-Bridge: %v", err)
	}
	defer response.Body.Close()
	body := new(bytes.Buffer)
	body.ReadFrom(response.Body)
	if response.StatusCode >= 400 {
		r, _ := regexp.Compile("<strong>Message:</strong>(.+)</div>")
		if matches := r.FindStringSubmatch(body.String()); len(matches) > 1 {
			return "", errors.NewLocalizedError("RSS-Bridge: %v", matches[1])
		}
		return "", errors.NewLocalizedError("RSS-Bridge: Server Failure (%d)", response.StatusCode)
	}
	return body.String(), nil
}
