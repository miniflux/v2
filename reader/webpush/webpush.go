package webpush

import (
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"log"
	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/sanitizer"
	"strings"
)

func SendPush(cfg *config.Config, subscriptions model.UserWebpushSubscriptions, entry *model.Entry, feed *model.Feed) {
	if "" == cfg.WebPushSubscriberEmail() || "" == cfg.WebPushVAPIDPublicKey() || "" == cfg.WebPushVAPIDPrivateKey() {
		return
	}

	logger.Debug("[WebPush:SendPush] Will send push notification for %s", entry.Title)

	for key := range subscriptions {
		// Decode subscription
		s := &webpush.Subscription{}
		err := json.Unmarshal([]byte(subscriptions[key].Subscription), s)

		if err != nil {
			log.Println(fmt.Errorf("invalid subscription: %v", err))
			continue
		}

		notification := new(Notification)
		notification.EntryTitle = entry.Title
		notification.EntryContent = sanitizeContent(entry)
		notification.FeedTitle = feed.Title

		// Send Notification
		notificationJSON := toJSON(notification)
		_, err = webpush.SendNotification([]byte(notificationJSON), s, &webpush.Options{
			Subscriber:      cfg.WebPushSubscriberEmail(),
			VAPIDPublicKey:  cfg.WebPushVAPIDPublicKey(),
			VAPIDPrivateKey: cfg.WebPushVAPIDPrivateKey(),
			TTL:             cfg.WebPushTTL(),
		})
		if err != nil {
			log.Println(fmt.Errorf("invalid subscription: %v", err))
			continue
		}
	}
}

func sanitizeContent(entry *model.Entry) string {
	return strings.TrimSpace(fmt.Sprintf("%.200s", sanitizer.StripTags(entry.Content))) + "..."
}

func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		logger.Error("[HTTP:JSON] %v", err)
		return []byte("")
	}

	return b
}

type Notification struct {
	FeedTitle    string `json:"feed_title"`
	EntryTitle   string `json:"entry_title"`
	EntryContent string `json:"entry_content"`
}
