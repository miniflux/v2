# Miniflux RSS and Atom feeds

Miniflux can expose your entries as regular RSS 2.0 and Atom 1.0 feeds that any
feed reader can subscribe to. This is intended as a lightweight alternative to
the Google Reader API: instead of replicating a stateful synchronization
protocol, Miniflux publishes read-only feed documents, with an optional
self-consuming inbox (the unread feed with `mark_read=1`).

## Endpoint

- RSS 2.0: `BASE_URL/rss/...`
- Atom 1.0: `BASE_URL/atom/...`

`BASE_URL` includes the Miniflux root URL and any configured `BasePath`.
Both formats expose exactly the same streams and accept the same parameters.

## Configuration

The feeds are disabled by default and are opt-in per user. Enable them from the
"Integrations" settings page (`BASE_URL/integrations`), in the "RSS / Atom"
section. Enabling them generates a random feed token, shown on the same page
together with ready-to-use feed URLs.

The feed token is separate from the REST API keys:

- It only grants access to the `/rss/` and `/atom/` endpoints. It cannot be used
  with the REST API, the Fever API, or the Google Reader API, and it cannot
  change account settings.
- It works whether or not the REST API is enabled (`DISABLE_API`).
- Disabling the feeds revokes the token immediately. Enabling them again
  generates a new one.
- The "Generate a new token" checkbox rotates the token, for example after a
  feed URL leaked. Existing feed URLs stop working.

## Authentication

The feed token can be provided in three ways, so that most feed readers work
without special support:

1. As a query parameter: `?token=<token>`
2. As the password of HTTP Basic authentication, with your Miniflux username as
   the username
3. As an `Authorization: Bearer <token>` header

Example:

```text
https://miniflux.example.com/rss/unread?token=0123456789abcdef
```

Security notes:

- Query parameters end up in proxy and web server access logs. Use HTTP Basic
  authentication when your reader supports it. Since the token only exposes the
  feeds, a leaked URL never gives access to the account; rotate the token from
  the Integrations page if it happens.
- The token is removed from the `atom:link rel="self"` URL generated in the
  feed, so the document itself never contains it.
- Tokens are matched by comparing SHA-256 digests, so the lookup does not leak
  the stored token through timing differences.
- The Atom feed `<id>` is built from the feed URL without its query string, so
  parameters such as `limit`, `mark_read`, or `q` never change the identity of
  the feed.
- Every request updates the "last login" timestamp of the user, like the Fever
  API does.

## Available feeds

| Path | Description |
| --- | --- |
| `/rss/unread` | Unread entries, newest first, excluding feeds and categories hidden from the Unread page. |
| `/rss/starred` | Starred entries. |
| `/rss/read` | Read entries (history). |
| `/rss/all` | All entries. |
| `/rss/feed/{feedID}` | Entries of one feed. |
| `/rss/category/{categoryID}` | Entries of one category. |
| `/rss/search?q=term` | Full-text search, same syntax as the search page. |

The same paths exist under `/atom/`, for example `/atom/starred`.

## Query parameters

| Parameter | Default | Description |
| --- | --- | --- |
| `token` | | Feed token from the Integrations page. |
| `limit` | `50` | Number of entries to return. Clamped to `1000`, like the REST API. |
| `mark_read` | `0` | Only for `/unread`. Set to `1` or `true` to mark the returned entries as read. |
| `q` | | Only for `/search`. Search query, required. |

## Notes on the unread feed

By default, `/rss/unread` is a read-only view, like every other feed: fetching
it never changes entry state. This matters because link previewers, prefetchers,
and crawlers can fetch a feed URL without any user action.

With `mark_read=1`, the unread feed behaves like an inbox: entries included in
the response are marked as read after the document has been rendered. This lets
simple feed readers act as a "process once" queue.

Consequences when `mark_read=1` is set:

- Only entries present in the response are marked, so a small `limit` never
  marks entries you have not received.
- A backlog larger than `limit` is drained over several requests, newest first:
  the next request returns the next batch, because the previous batch is now
  read. A reader that wants everything at once can ask for `limit=1000`.
- `HEAD` requests never mark entries as read, because some readers probe feeds
  with `HEAD` before downloading them.
- Entries that were already read are never included again, so a reader that
  fails to store a response loses those entries.
- Entries from feeds and categories hidden globally are neither listed nor
  marked as read, matching the Unread page.

The other feeds never change entry state, even with `mark_read=1`.

## Per-entry metadata

Each entry includes:

- A stable identifier built from the Miniflux entry ID
  (`urn:miniflux:entry:<id>`), also used as the RSS GUID.
- Title, link, publication date, and author.
- The full HTML content, escaped, with media rewritten through the media proxy
  when it is enabled. RSS items carry it in `<description>` and
  `<content:encoded>`; Atom entries carry it once, in `<content>`.
- The feed title and category as categories/keywords.
- Enclosures as `<enclosure>` (RSS) or `rel="enclosure"` links (Atom). Their
  URLs go through the media proxy, according to `MEDIA_PROXY_MODE` and
  `MEDIA_PROXY_RESOURCE_TYPES`, like the REST and Google Reader APIs.
- The comment URL, when the feed provides one (RSS only).

## Response and caching

- `Content-Type: application/rss+xml; charset=utf-8` or
  `application/atom+xml; charset=utf-8`.
- `Cache-Control: private, no-store`, because the documents are authenticated
  and the unread feed can have side effects.
- Unknown feeds return `404`, authentication failures return `401` with a
  `WWW-Authenticate: Basic` challenge, and invalid parameters return `400`.
- `GET /rss/` and `GET /atom/` return a short plain-text help page listing the
  available endpoints (authentication required).
