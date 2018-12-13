package media // import "miniflux.app/reader/media"

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"miniflux.app/crypto"
	"miniflux.app/http/client"
	"miniflux.app/logger"
	"miniflux.app/model"
)

var queries = []string{
	"img[src]",
}

// URLHash returns the hash of a media url
func URLHash(mediaURL string) string {
	return crypto.Hash(strings.Trim(mediaURL, " "))
}

// FindMedia try to find the media cache of the URL.
func FindMedia(mediaURL string) (*model.Media, error) {
	if strings.HasPrefix(mediaURL, "data:") {
		return nil, fmt.Errorf("refuse to cache 'data' scheme media")
	}

	logger.Debug("[FindMedia] Fetching media => %s", mediaURL)
	media, err := downloadMedia(mediaURL)
	if err != nil {
		return nil, err
	}

	return media, nil
}

func downloadMedia(mediaURL string) (*model.Media, error) {
	clt := client.New(mediaURL)
	response, err := clt.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to download mediaURL: %v", err)
	}

	if response.HasServerFailure() {
		return nil, fmt.Errorf("unable to download media: status=%d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read downloaded media: %v", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("downloaded media is empty, mediaURL=%s", mediaURL)
	}

	media := &model.Media{
		UrlHash:  URLHash(mediaURL),
		MimeType: response.ContentType,
		Content:  body,
	}

	return media, nil
}

// ParseDocument parse the entry content and returns media urls of it
func ParseDocument(entry *model.Entry) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entry.Content))
	if err != nil {
		return nil, fmt.Errorf("unable to read document: %v", err)
	}

	urls := make([]string, 0)
	for _, query := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			href := strings.Trim(s.AttrOr("src", ""), " ")
			if href == "" || strings.HasPrefix(href, "data:") {
				return
			}
			href, err = url.AbsoluteURL(entry.URL, href)
			if err != nil {
				return
			}
			urls = append(urls, href)
		})
	}

	return urls, nil
}

// RedirectMedia redirects media urls in entryContent to media cache if exists.
func RedirectMedia(entry *model.Entry, medias map[string]*model.Media, baseURL string) {
	if len(medias) == 0 {
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entry.Content))
	if err != nil {
		log.Println(err)
		return
	}
	for _, q := range queries {
		matches := doc.Find(q)
		if matches.Length() == 0 {
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		matches.Each(func(i int, img *goquery.Selection) {
			href := img.AttrOr("src", "")
			if href == "" || strings.HasPrefix(href, "data:") {
				return
			}
			href, err = url.AbsoluteURL(entry.URL, href)
			if err != nil {
				return
			}
			hash := URLHash(href)
			if _, ok := medias[hash]; !ok {
				return
			}

			u, err := url.AbsoluteURL(baseURL, "/media/"+hash)
			if err != nil {
				log.Fatal(err)
			}
			img.SetAttr("src", u)
		})

		entry.Content, _ = doc.Find("body").First().Html()

	}
}
