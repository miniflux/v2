# Miniflux Fever API

This document describes the Fever-compatible API implemented by the `internal/fever` package in this repository.

## Endpoint

- Path: `BASE_URL/fever/`
- Methods: not restricted by the router; read requests are typically sent as `GET`, write requests should be sent as `POST`
- Response format: JSON only
- Reported API version: `3`

## Authentication

Fever authentication is enabled per user from the Miniflux integrations page.

- `Fever Username` and `Fever Password` are configured in Miniflux
- Miniflux stores the Fever token as the MD5 hash of `username:password`
- Clients authenticate by sending that token as the `api_key` parameter
- Token lookup is case-insensitive

Example:

```text
api_key = md5("fever_username:fever_password")
```

Example shell command:

```bash
printf '%s' 'fever_username:fever_password' | md5sum
```

Authentication failure does not return HTTP 401. The middleware returns HTTP 200 with:

```json
{
  "api_version": 3,
  "auth": 0
}
```

On successful authentication, every response includes:

- `api_version`: always `3`
- `auth`: always `1`
- `last_refreshed_on_time`: current server Unix timestamp at response time

## Dispatch Rules

The handler selects the first matching operation in this order:

1. `groups`
2. `feeds`
3. `favicons`
4. `unread_item_ids`
5. `saved_item_ids`
6. `items`
7. `mark=item`
8. `mark=feed`
9. `mark=group`

If no selector is provided, the server returns the base authenticated response only.

For read operations, the selector must be present in the query string. For write operations, `mark`, `as`, `id`, and `before` are read from request form values, so they may come from the query string or a form body.

## Read Operations

### `?groups`

Returns:

- `groups`: list of categories
- `feeds_groups`: mapping of category IDs to feed IDs

Response shape:

```json
{
  "api_version": 3,
  "auth": 1,
  "last_refreshed_on_time": 1710000000,
  "groups": [
    {
      "id": 1,
      "title": "All"
    }
  ],
  "feeds_groups": [
    {
      "group_id": 1,
      "feed_ids": "10,11"
    }
  ]
}
```

Notes:

- `groups` are Miniflux categories
- `feeds_groups.feed_ids` is a comma-separated string
- categories with no feeds are returned in `groups` but have no `feeds_groups` entry

### `?feeds`

Returns:

- `feeds`: list of feeds
- `feeds_groups`: mapping of category IDs to feed IDs

Feed fields:

- `id`
- `favicon_id`
- `title`
- `url`
- `site_url`
- `is_spark`
- `last_updated_on_time`

Notes:

- `favicon_id` is `0` when the feed has no icon
- `is_spark` is always `0` in this implementation
- `last_updated_on_time` is the feed check time as a Unix timestamp

### `?favicons`

Returns:

- `favicons`: list of favicon objects

Favicon fields:

- `id`
- `data`

Notes:

- `data` is a data URL such as `image/png;base64,...`

### `?unread_item_ids`

Returns:

- `unread_item_ids`: comma-separated list of unread entry IDs

Response shape:

```json
{
  "api_version": 3,
  "auth": 1,
  "last_refreshed_on_time": 1710000000,
  "unread_item_ids": "100,101,102"
}
```

### `?saved_item_ids`

Returns:

- `saved_item_ids`: comma-separated list of starred entry IDs

### `?items`

Returns:

- `items`: list of entries
- `total_items`: total number of non-removed entries for the user

Item fields:

- `id`
- `feed_id`
- `title`
- `author`
- `html`
- `url`
- `is_saved`
- `is_read`
- `created_on_time`

The implementation always excludes entries whose status is `removed`.

#### Pagination and filtering

The handler applies a fixed limit of 50 items.

Supported parameters:

- `since_id`: when greater than `0`, returns entries with `id > since_id`, ordered by `id ASC`
- `max_id`: when equal to `0`, returns the most recent entries ordered by `id DESC`; when greater than `0`, returns entries with `id < max_id`, ordered by `id DESC`
- `with_ids`: comma-separated list of entry IDs to fetch

Selector precedence inside `?items` is:

1. `since_id`
2. `max_id`
3. `with_ids`
4. no item filter

Notes:

- `with_ids` does not enforce the 50-ID maximum mentioned in older Fever documentation
- invalid `with_ids` members are parsed as `0` and do not match normal entries
- when `items` is requested without `since_id`, `max_id`, or `with_ids`, the code applies no explicit `ORDER BY`, so result ordering is not guaranteed by SQL
- `html` is returned after Miniflux content rewriting and may include media-proxy-rewritten URLs

Example:

```json
{
  "api_version": 3,
  "auth": 1,
  "last_refreshed_on_time": 1710000000,
  "total_items": 245,
  "items": [
    {
      "id": 100,
      "feed_id": 10,
      "title": "Example entry",
      "author": "Author",
      "html": "<p>Content</p>",
      "url": "https://example.org/post",
      "is_saved": 0,
      "is_read": 1,
      "created_on_time": 1709990000
    }
  ]
}
```

## Write Operations

Normal successful write operations return the base authenticated response:

```json
{
  "api_version": 3,
  "auth": 1,
  "last_refreshed_on_time": 1710000000
}
```

### `mark=item`

Parameters:

- `mark=item`
- `id=<entry_id>`
- `as=read|unread|saved|unsaved`

Behavior:

- `as=read`: marks the entry as read
- `as=unread`: marks the entry as unread
- `as=saved`: toggles the starred flag
- `as=unsaved`: toggles the starred flag

Important:

- `saved` and `unsaved` both call the same toggle operation
- sending `as=saved` twice will save, then unsave
- sending `as=unsaved` twice will unsave, then save
- if `id <= 0`, the handler returns without writing a response body
- if the entry does not exist or is already removed, the server returns the base response without an error

### `mark=feed`

Parameters:

- `mark=feed`
- `as=read`
- `id=<feed_id>`
- `before=<unix_timestamp>`

Behavior:

- marks unread entries in the feed as read when `published_at < before`
- the update runs asynchronously in a goroutine after the response is returned

Notes:

- if `id <= 0`, the handler returns without writing a response body
- if `before` is missing or invalid, it is treated as Unix time `0`, which usually means nothing is marked as read

### `mark=group`

Parameters:

- `mark=group`
- `as=read`
- `id=<group_id>`
- `before=<unix_timestamp>`

Behavior:

- `id=0`: marks all unread entries as read, ignoring `before`
- `id>0`: marks unread entries in the matching category as read when `published_at < before`
- the update runs asynchronously in a goroutine after the response is returned

Notes:

- group IDs map to Miniflux category IDs
- if `id < 0`, the handler returns without writing a response body
- if `before` is missing or invalid for `id>0`, it is treated as Unix time `0`, which usually means nothing is marked as read

## Error Handling

Authentication failures:

- HTTP status: `200`
- body: `{"api_version":3,"auth":0}`

Internal errors:

- HTTP status: `500`
- body:

```json
{
  "error_message": "..."
}
```

## Differences From Generic Fever Documentation

This implementation is Fever-compatible, but it does not match every detail of historical Fever API docs.

- Responses are always JSON; `api=xml` is mentioned in code comments but is not implemented
- `api_version` is `3`
- `last_refreshed_on_time` is set to the current response time, not the timestamp of the most recently refreshed feed
- the `Kindling` and `Sparks` super groups are not returned
- `feeds[].is_spark` is always `0`
- item ordering without explicit pagination parameters is unspecified
- `as=saved` and `as=unsaved` toggle the saved flag instead of setting it absolutely

## Examples

Fetch groups:

```bash
curl -s 'https://miniflux.example.com/fever/?api_key=TOKEN&groups'
```

Fetch most recent items:

```bash
curl -s 'https://miniflux.example.com/fever/?api_key=TOKEN&items&max_id=0'
```

Fetch items after a known ID:

```bash
curl -s 'https://miniflux.example.com/fever/?api_key=TOKEN&items&since_id=123'
```

Mark an item as read:

```bash
curl -s -X POST 'https://miniflux.example.com/fever/' \
  -d 'api_key=TOKEN' \
  -d 'mark=item' \
  -d 'as=read' \
  -d 'id=123'
```

Mark a feed as read before a timestamp:

```bash
curl -s -X POST 'https://miniflux.example.com/fever/' \
  -d 'api_key=TOKEN' \
  -d 'mark=feed' \
  -d 'as=read' \
  -d 'id=10' \
  -d 'before=1710000000'
```

Mark all items as read through the group endpoint:

```bash
curl -s -X POST 'https://miniflux.example.com/fever/' \
  -d 'api_key=TOKEN' \
  -d 'mark=group' \
  -d 'as=read' \
  -d 'id=0'
```
