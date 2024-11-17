package nostr

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/nbd-wtf/go-nostr/sdk"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/storage"
)

var (
	NostrSdk *sdk.System
)

func GetIcon(feed *model.Feed) (bool, string) {
	yes, profile := IsItNostr(feed.FeedURL)

	if yes {
		return true, profile.Picture
	}

	return false, ""
}

func CreateFeed(store *storage.Storage, user *model.User, feedCreationRequest *model.FeedCreationRequest) (bool, *model.Feed) {
	ctx := context.Background()
	yes, profile := IsItNostr(feedCreationRequest.FeedURL)

	if yes {
		subscription := &model.Feed{}
		nprofile := profile.Nprofile(ctx, NostrSdk, 3)
		subscription.Title = profile.Name
		subscription.UserID = user.ID
		subscription.UserAgent = feedCreationRequest.UserAgent
		subscription.Cookie = feedCreationRequest.Cookie
		subscription.Username = feedCreationRequest.Username
		subscription.Password = feedCreationRequest.Password
		subscription.Crawler = feedCreationRequest.Crawler
		subscription.FetchViaProxy = feedCreationRequest.FetchViaProxy
		subscription.HideGlobally = feedCreationRequest.HideGlobally
		subscription.FeedURL = fmt.Sprintf("nostr:%s", nprofile)
		subscription.SiteURL = fmt.Sprintf("nostr:%s", nprofile)
		subscription.WithCategoryID(feedCreationRequest.CategoryID)
		subscription.CheckedNow()

		if storeErr := store.CreateFeed(subscription); storeErr != nil {
			return false, nil
		}

		if err := RefreshFeed(store, user, subscription); !err {
			// TODO: error handling
			return false, nil
		}

		return true, subscription
	}

	return false, nil
}

func Initialize() {
	NostrSdk = sdk.NewSystem(
		sdk.WithRelayListRelays([]string{
			"wss://nos.lol", "wss://nostr.mom", "wss://nostr.bitcoiner.social", "wss://relay.damus.io", "wss://nostr-pub.wellorder.net"}, // some standard relays
		),
	)
}

func RefreshFeed(store *storage.Storage, user *model.User, originalFeed *model.Feed) bool {
	ctx := context.Background()
	if yes, profile := IsItNostr(originalFeed.FeedURL); yes {
		relays := NostrSdk.FetchOutboxRelays(ctx, profile.PubKey, 3)
		evchan := NostrSdk.Pool.SubManyEose(ctx, relays, nostr.Filters{
			{
				Authors: []string{profile.PubKey},
				Kinds:   []int{nostr.KindArticle},
				Limit:   32,
			},
		})
		updatedFeed := originalFeed
		for event := range evchan {

			publishedAt := event.CreatedAt.Time()
			if publishedAtTag := event.Tags.GetFirst([]string{"published_at"}); publishedAtTag != nil && len(*publishedAtTag) >= 2 {
				i, err := strconv.ParseInt((*publishedAtTag)[1], 10, 64)
				if err != nil {
					publishedAt = time.Unix(i, 0)
				}
			}

			naddr, err := nip19.EncodeEntity(event.PubKey, event.Kind, event.Tags.GetD(), relays)
			if err != nil {
				continue
			}

			title := ""
			titleTag := event.Tags.GetFirst([]string{"title"})
			if titleTag != nil && len(*titleTag) >= 2 {
				title = (*titleTag)[1]
			}

			// format content from markdown to html
			entry := &model.Entry{
				Date:    publishedAt,
				Title:   title,
				Content: event.Content,
				URL:     fmt.Sprintf("nostr:%s", naddr),
				Hash:    fmt.Sprintf("nostr:%s:%s", event.PubKey, event.Tags.GetD()),
			}

			rewrite.Rewriter(entry.URL, entry, "parse_markdown")

			updatedFeed.Entries = append(updatedFeed.Entries, entry)

		}

		processor.ProcessFeedEntries(store, updatedFeed, user, true)

		_, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, updatedFeed.Entries, false)
		if storeErr != nil {
			// TODO: Error handling
			return false
		}

		return true
	}
	return false
}

func IsItNostr(url string) (bool, *sdk.ProfileMetadata) {
	ctx := context.Background()
	if NostrSdk == nil {
		Initialize()
	}

	// check for nostr url prefixes
	if strings.HasPrefix(url, "nostr://") {
		url = url[8:]
	} else if strings.HasPrefix(url, "nostr:") {
		url = url[6:]
	} else {
		// only accept nostr: or nostr:// urls for now
		return false, nil
	}

	// check for npub or nprofile
	if prefix, _, err := nip19.Decode(url); err == nil {
		if prefix == "nprofile" || prefix == "npub" {
			profile, err := NostrSdk.FetchProfileFromInput(ctx, url)
			if err != nil {
				return false, nil
			}
			return true, &profile
		}
	}

	// only do nip05 check when nostr prefix
	if nip05.IsValidIdentifier(url) {
		profile, err := NostrSdk.FetchProfileFromInput(ctx, url)
		if err != nil {
			return false, nil
		}
		return true, &profile
	}

	return false, nil

}
