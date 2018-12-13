package ui // import "miniflux.app/cache"

import (
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/model"

	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
)

func (h *handler) provideCache(w http.ResponseWriter, r *http.Request) {
	urlHash := request.RouteStringParam(r, "urlHash")
	media := model.Media{UrlHash: urlHash}
	err := h.store.MediaByHash(&media)
	if err != nil || media.ID == 0 {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(urlHash, 48*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", media.MimeType)
		b.WithBody(media.Content)
		b.Write()
	})
}
