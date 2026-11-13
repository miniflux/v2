# Miniflux RSS and Atom feeds

Miniflux can expose your entries as regular RSS 2.0 and Atom 1.0 feeds that any
feed reader can subscribe to. This is intended as a lightweight alternative to
the Google Reader API: instead of replicating a stateful synchronization
protocol, Miniflux publishes read-only feed documents plus one self-consuming
inbox (the unread feed).

## Endpoint

- RSS 2.0: `BASE_URL/rss/...`
- Atom 1.0: `BASE_URL/atom/...`

`BASE_URL` includes the Miniflux root URL and any configured `BasePath`.
Both formats expose exactly the same streams and accept the same parameters.

## Configuration

There is nothing to enable: the feeds are mounted unconditionally, there is no
setting, no password and no database migration. They reuse the API keys that
already exist for the REST API and the Google Reader API, so any reader can
subscribe with the same key it uses for those integrations.

API keys are created from the "API Keys" settings page (`BASE_URL/keys`), which
requires the REST API to be enabled (`DISABLE_API` unset).

## Authentication

The feeds are protected with a Miniflux API key.

The API key can be provided in three ways, so that most feed readers work
without special support:

1. As a query parameter: `?api_key=<key>`
2. As the password of HTTP Basic authentication, with your Miniflux username as
   the username
3. As an `Authorization: Bearer <key>` header

Example:

```text
https://miniflux.example.com/rss/unread?api_key=0123456789abcdef
```

Security notes:

- Query parameters end up in proxy and web server access logs. Use HTTP Basic
  authentication when your reader supports it.
- The API key is removed from the `atom:link rel="self"` URL generated in the
  feed, so the document itself never contains the key.
- The Atom feed `<id>` is built from the feed URL without its query string, so
  parameters such as `limit`, `mark_read`, or `q` never change the identity of
  the feed.
- Every request updates the "last login" and "API key last used" timestamps of
  the user, like the REST API does.

## Available feeds

| Path | Description |
| --- | --- |
| `/rss/unread` | Unread entries, newest first. Returned entries are marked as read. |
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
| `api_key` | | Miniflux API key. |
| `limit` | `50` | Number of entries to return. Clamped to `1000`, like the REST API. |
| `mark_read` | `1` | Only for `/unread`. Set to `0` or `false` to keep entries unread. |
| `q` | | Only for `/search`. Search query, required. |

## Notes on the unread feed

`/rss/unread` behaves like an inbox: entries included in the response are marked
as read after the document has been rendered. This matches the behavior you get
by reading entries in the web interface, and it lets simple feed readers act as a
"process once" queue.

Consequences:

- Only entries present in the response are marked, so a small `limit` never
  marks entries you have not received.
- A backlog larger than `limit` is drained over several requests, newest first:
  the next request returns the next batch, because the previous batch is now
  read. A reader that wants everything at once can ask for `limit=1000`.
- `HEAD` requests never mark entries as read, because some readers probe feeds
  with `HEAD` before downloading them.
- Entries that were already read are never included, so a reader that fails to
  store a response loses those entries. Subscribe to `/rss/unread?mark_read=0`
  if you want a non-destructive unread feed.

The other feeds never change entry state.

## Per-entry metadata

Each entry includes:

- A stable identifier built from the Miniflux entry ID
  (`urn:miniflux:entry/<id>`), also used as the RSS GUID.
- Title, link, publication date, and author.
- The full HTML content, escaped. RSS documents also include it in
  `<content:encoded>`.
- The feed title and category as categories/keywords.
- Enclosures as `<enclosure>` (RSS) or `rel="enclosure"` links (Atom).
- The comment URL, when the feed provides one (RSS only).

## Response and caching

- `Content-Type: application/rss+xml; charset=utf-8` or
  `application/atom+xml; charset=utf-8`.
- `Cache-Control: private, no-store`, because the documents are authenticated
  and the unread feed has side effects.
- Unknown feeds return `404`, authentication failures return `401` with a
  `WWW-Authenticate: Basic` challenge, and invalid parameters return `400`.
- `GET /rss/` and `GET /atom/` return a short plain-text help page listing the
  available endpoints.
