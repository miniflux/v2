# Miniflux Google Reader API

This document describes the Google Reader compatible API implemented by the `internal/googlereader` package in this repository.

Miniflux implements a compatibility subset intended for existing Google Reader clients. It is not a full reimplementation of the historical Google Reader API, and several behaviors are intentionally narrower or implementation-specific.

## Endpoint

- Client login path: `BASE_URL/accounts/ClientLogin`
- API prefix: `BASE_URL/reader/api/0`
- `BASE_URL` includes the Miniflux root URL and any configured `BasePath`
- Response format:
  - `ClientLogin`: plain text by default, JSON when `output=json`
  - most API reads: JSON
  - most API writes: plain text `OK`

## Enabling the API

Google Reader compatibility is configured per user from the Miniflux integrations page.

- `Google Reader API` must be enabled
- `Google Reader Username` must be unique across all Miniflux users
- `Google Reader Password` is stored as a bcrypt hash

The Google Reader username and password are separate integration credentials. They are not the Miniflux account password.

## Authentication

### `POST /accounts/ClientLogin`

This endpoint exchanges the configured Google Reader username and password for an auth token.

Form parameters:

- `Email`: Google Reader username
- `Passwd`: Google Reader password
- `output`: optional, set to `json` for a JSON response

Successful responses:

- default: plain text
- with `output=json`: JSON

Example plain-text response:

```text
SID=readeruser/0123456789abcdef...
LSID=readeruser/0123456789abcdef...
Auth=readeruser/0123456789abcdef...
```

Example JSON response:

```json
{
  "SID": "readeruser/0123456789abcdef...",
  "LSID": "readeruser/0123456789abcdef...",
  "Auth": "readeruser/0123456789abcdef..."
}
```

On authentication failure, `ClientLogin` returns HTTP `401` with the normal JSON error body:

```json
{
  "error_message": "access unauthorized"
}
```

### Auth token format

The token format is:

```text
<googlereader_username>/<hex_digest>
```

The digest is generated server-side from:

- the Google Reader username
- the stored bcrypt hash of the Google Reader password

Specifically, the code computes an HMAC-SHA256 digest of an empty message using the key:

```text
googlereader_username + bcrypt_hash
```

Because the bcrypt hash is only known to the server, clients should not try to precompute the token. Use `ClientLogin` or `GET /reader/api/0/token`.

### Authenticating API calls

Miniflux uses different auth mechanisms for `GET` and `POST` requests:

- `GET` requests must send the header `Authorization: GoogleLogin auth=<token>`
- `POST` requests are authenticated with `T=<token>` read from the parsed form values

Notes:

- the auth scheme must be exactly `GoogleLogin`
- the auth field name must be exactly lowercase `auth`
- for `POST`, `T` may come from the URL query or the form body because the server reads merged form values
- `POST` requests do not accept the token from the `Authorization` header
- `GET` requests do not accept the token from the query string

### `GET /reader/api/0/token`

This endpoint requires normal `GET` authentication and returns the same token as plain text.

Many Google Reader clients use this as the edit token for subsequent write requests. In Miniflux, the edit token and auth token are the same value.

### Authentication failure on `/reader/api/0/*`

When API authentication fails under `/reader/api/0`, Miniflux returns:

- HTTP `401`
- header `X-Reader-Google-Bad-Token: true`
- content type `text/plain; charset=utf-8`
- body `Unauthorized`

This is different from `ClientLogin`, which returns a JSON `401`.

## Identifier formats

### Stream IDs

The implementation recognizes these stream forms:

- built-in streams:
  - `user/-/state/com.google/read`
  - `user/-/state/com.google/starred`
  - `user/-/state/com.google/reading-list`
  - `user/-/state/com.google/kept-unread`
  - `user/-/state/com.google/broadcast`
  - `user/-/state/com.google/broadcast-friends`
  - `user/-/state/com.google/like`
- user-specific equivalents:
  - `user/<user_id>/state/com.google/...`
- label streams:
  - `user/-/label/<name>`
  - `user/<user_id>/label/<name>`
- feed streams:
  - `feed/<value>`

Important feed stream difference:

- read APIs usually emit `feed/<numeric_feed_id>`
- `subscription/edit` with `ac=subscribe` expects `feed/<absolute_feed_url>`
- `subscription/edit` with `ac=edit` or `ac=unsubscribe` expects `feed/<numeric_feed_id>`

So `feed/<...>` is not a single stable identifier format across all endpoints.

### Item IDs

`edit-tag` and `stream/items/contents` accept repeated `i` parameters in all of these formats:

- long Google Reader form: `tag:google.com,2005:reader/item/00000000148b9369`
- short prefixed hexadecimal form: `tag:google.com,2005:reader/item/2f2`
- bare 16-character hexadecimal form: `000000000000048c`
- decimal entry ID: `12345`

Responses use different forms depending on endpoint:

- `stream/items/ids` returns decimal IDs as strings
- `stream/items/contents` returns long-form Google Reader item IDs

## Common response conventions

JSON errors use this shape:

```json
{
  "error_message": "..."
}
```

Plain-text success responses from write endpoints are usually:

```text
OK
```

## POST parameter parsing

Most `POST` handlers call `ParseForm()` and read from `r.Form`, so parameters may be supplied either in the query string or in a standard form body.

Important exception:

- `POST /reader/api/0/edit-tag` reads `a` and `r` from `r.PostForm`, so those tag lists must come from the request body

Because `GET` auth comes only from the `Authorization` header, query parameters never authenticate `GET` requests even when other parameters are read from the query string.

## Endpoint reference

### `GET /reader/api/0/user-info`

Returns JSON only. No `output=json` parameter is required.

Response fields:

- `userId`: Miniflux user ID as a string
- `userName`: Miniflux username
- `userProfileId`: same value as `userId`
- `userEmail`: same value as `userName`

Example:

```json
{
  "userId": "1",
  "userName": "demo",
  "userProfileId": "1",
  "userEmail": "demo"
}
```

### `GET /reader/api/0/tag/list?output=json`

Returns the starred state and user labels.

Notes:

- `output=json` is required
- only labels and the starred state are returned
- built-in states such as `read` and `reading-list` are not listed here

Response shape:

```json
{
  "tags": [
    {
      "id": "user/1/state/com.google/starred"
    },
    {
      "id": "user/1/label/Tech",
      "label": "Tech",
      "type": "folder"
    }
  ]
}
```

### `GET /reader/api/0/subscription/list?output=json`

Returns the user's feeds.

Notes:

- `output=json` is required
- each feed is reported with a numeric feed stream ID such as `feed/42`
- `categories` always contains the Miniflux category as a Google Reader folder

Response shape:

```json
{
  "subscriptions": [
    {
      "id": "feed/42",
      "title": "Example Feed",
      "categories": [
        {
          "id": "user/1/label/Tech",
          "label": "Tech",
          "type": "folder"
        }
      ],
      "url": "https://example.org/feed.xml",
      "htmlUrl": "https://example.org/",
      "iconUrl": "https://miniflux.example.com/icon/..."
    }
  ]
}
```

### `POST /reader/api/0/subscription/quickadd`

Subscribes to the first discovered feed for the given absolute URL.

Form parameters:

- `T`: auth token
- `quickadd`: absolute URL

Response shape when a feed is found:

```json
{
  "numResults": 1,
  "query": "https://example.org/feed.xml",
  "streamId": "feed/42",
  "streamName": "Example Feed"
}
```

Response shape when no feed is found:

```json
{
  "numResults": 0
}
```

Notes:

- the request URL must be absolute
- the created subscription is assigned to the user's first category when no explicit category is provided

### `POST /reader/api/0/subscription/edit`

Edits subscriptions. Successful requests return plain text `OK`.

Form parameters:

- `T`: auth token
- `ac`: action
- `s`: repeated stream ID
- `a`: optional destination label stream
- `t`: optional title

Supported actions:

- `ac=subscribe`
- `ac=unsubscribe`
- `ac=edit`

Behavior by action:

- `subscribe`
  - only the first `s` value is used
  - `s` must be `feed/<absolute_feed_url>`
  - `a`, when present, must be a label stream
  - `t`, when present, becomes the feed title after creation
- `unsubscribe`
  - every `s` must be `feed/<numeric_feed_id>`
- `edit`
  - only the first `s` value is used
  - `s` must be `feed/<numeric_feed_id>`
  - `t` renames the feed
  - `a` moves the feed to a label, and must be a label stream

Notable limitations:

- removing a label is not implemented here
- `subscribe`, `edit`, and `unsubscribe` do not share the same feed ID format

### `POST /reader/api/0/rename-tag`

Renames a label. Successful requests return plain text `OK`.

Form parameters:

- `T`: auth token
- `s`: source label stream
- `dest`: destination label stream

Rules:

- both `s` and `dest` must be label streams
- the destination label name must not be empty
- if the source label does not exist, the endpoint returns HTTP `404`

### `POST /reader/api/0/disable-tag`

Deletes one or more labels and reassigns affected feeds to the user's first remaining category.

Form parameters:

- `T`: auth token
- `s`: repeated label stream

Rules:

- only label streams are supported
- at least one category must remain after deletion, otherwise the operation fails

Successful requests return plain text `OK`.

### `POST /reader/api/0/edit-tag`

Marks entries read or unread and starred or unstarred.

Form parameters:

- `T`: auth token
- `i`: repeated item ID
- `a`: repeated tag stream to add
- `r`: repeated tag stream to remove

Supported tag semantics:

- add `user/.../state/com.google/read`: mark read
- remove `user/.../state/com.google/read`: mark unread
- add `user/.../state/com.google/kept-unread`: mark unread
- remove `user/.../state/com.google/kept-unread`: mark read
- add `user/.../state/com.google/starred`: star
- remove `user/.../state/com.google/starred`: unstar

Special cases:

- `read` and `kept-unread` cannot be combined in conflicting ways in the same request
- `starred` cannot be present in both add and remove
- `broadcast` and `like` are recognized but ignored
- unsupported tag types cause an error

Successful requests return plain text `OK`.

### `GET /reader/api/0/stream/items/ids?output=json`

Returns item IDs for one stream.

Required query parameters:

- `output=json`
- `s=<stream_id>`

Optional query parameters:

- `n`: maximum number of items to return
- `c`: numeric offset continuation token
- `r`: sort direction, `o` for ascending, anything else for descending
- `ot`: only items published after this Unix timestamp in seconds
- `nt`: only items published before this Unix timestamp in seconds
- `xt`: repeated exclude target stream
- `it`: repeated filter target stream, parsed but currently ignored

Supported `s` values:

- `user/.../state/com.google/reading-list`
- `user/.../state/com.google/starred`
- `user/.../state/com.google/read`
- `feed/<numeric_feed_id>`

Notes:

- exactly one `s` value is expected
- label streams are not supported here
- when `xt` contains the `read` stream, `reading-list` and `feed/<id>` behave as unread-only queries
- if `n` is omitted, the query is effectively unbounded
- `continuation` is a numeric offset encoded as a JSON string, not an opaque token

Response shape:

```json
{
  "itemRefs": [
    {
      "id": "12345"
    },
    {
      "id": "12344"
    }
  ],
  "continuation": "2"
}
```

### `POST /reader/api/0/stream/items/contents`

Returns content for specific items.

Required parameters:

- `T`: auth token
- `output=json`
- `i`: repeated item ID

Optional query parameters:

- `r`: sort direction, `o` for ascending, anything else for descending

Implementation notes:

- the route is `POST` only
- `T`, `output`, and `i` are read from merged form values, so they may be supplied in the query string or the form body
- the handler parses stream filter query parameters, but in practice only the sort direction affects the result

Response shape:

```json
{
  "direction": "ltr",
  "id": "user/-/state/com.google/reading-list",
  "title": "Reading List",
  "self": [
    {
      "href": "https://miniflux.example.com/reader/api/0/stream/items/contents"
    }
  ],
  "updated": 1710000000,
  "author": "demo",
  "items": [
    {
      "id": "tag:google.com,2005:reader/item/00000000148b9369",
      "categories": [
        "user/1/state/com.google/reading-list",
        "user/1/label/Tech",
        "user/1/state/com.google/starred"
      ],
      "title": "Example entry",
      "crawlTimeMsec": "1710000000123",
      "timestampUsec": "1710000000123456",
      "published": 1710000000,
      "updated": 1710000300,
      "author": "Author",
      "alternate": [
        {
          "href": "https://example.org/post",
          "type": "text/html"
        }
      ],
      "summary": {
        "direction": "ltr",
        "content": "<p>Content</p>"
      },
      "content": {
        "direction": "ltr",
        "content": "<p>Content</p>"
      },
      "origin": {
        "streamId": "feed/42",
        "title": "Example Feed",
        "htmlUrl": "https://example.org/"
      },
      "enclosure": [],
      "canonical": [
        {
          "href": "https://example.org/post"
        }
      ]
    }
  ]
}
```

Notes:

- top-level `id` and `title` are hard-coded as the reading list
- `summary.content` and `content.content` both contain the rewritten entry content
- enclosure URLs and embedded media may be rewritten through the Miniflux media proxy

### `POST /reader/api/0/mark-all-as-read`

Marks items as read before a timestamp. Successful requests return plain text `OK`.

Form parameters:

- `T`: auth token
- `s`: stream ID
- `ts`: optional timestamp

Supported `s` values:

- `feed/<numeric_feed_id>`
- `user/.../label/<name>`
- `user/.../state/com.google/reading-list`

Timestamp handling:

- if `ts` has at least 16 digits, it is interpreted as microseconds since the Unix epoch
- otherwise it is interpreted as seconds since the Unix epoch
- if `ts` is omitted, Miniflux uses the current server time

Notes:

- only unread entries published before `ts` are marked as read
- unsupported stream types are effectively a no-op and still return `OK`

### Catch-all unimplemented endpoints

Any other `GET` or `POST` path under `/reader/api/0/` is caught by the fallback handler and returns:

```json
[]
```

with HTTP `200`.

## Compatibility notes and deviations

These differences are important for client authors:

- only a subset of Google Reader endpoints is implemented
- feed stream IDs are numeric in read responses, but `ac=subscribe` expects `feed/<absolute_feed_url>`
- `stream/items/ids` returns decimal entry IDs, while `stream/items/contents` returns long-form Google Reader item IDs
- pagination uses `c` as a numeric SQL offset, not an opaque continuation token
- `it` filter targets are parsed but currently ignored
- `tag/list` returns only `starred` and user labels
- API auth failures under `/reader/api/0/*` return plain text `401 Unauthorized`, not JSON
- unknown `/reader/api/0/*` endpoints return `[]` with `200`, not `404`
