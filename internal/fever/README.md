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
- Clients authenticate by sending that token as the `api_key` parameter, in the query string or the form body
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
  "auth": 0,
  "last_refreshed_on_time": 0
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

Lists are always JSON arrays: they are empty, not `null`, when there is nothing to return.

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
- `total_items`: total number of entries of the user, whatever the selection

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

#### Pagination and filtering

The handler applies a fixed limit of 50 items.

Supported parameters:

- `since_id`: returns entries with `id > since_id` when greater than `0`, and the oldest entries otherwise, ordered by `id ASC`
- `max_id`: returns entries with `id < max_id` when greater than `0`, and the most recent entries otherwise (including `0`), ordered by `id DESC`
- `with_ids`: comma-separated list of entry IDs to fetch, ordered by `id ASC`

Selector precedence inside `?items` is:

1. `since_id`
2. `max_id`
3. `with_ids`
4. no item filter

Notes:

- `with_ids` does not enforce the 50-ID maximum mentioned in older Fever documentation, but the 50-row limit still applies: a longer list returns its 50 lowest IDs
- invalid and non-positive `with_ids` members are ignored, and a list without valid IDs returns no items
- when `items` is requested without `since_id`, `max_id`, or `with_ids`, the 50 oldest entries are returned, ordered by `id ASC`
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
- `as=saved`: stars the entry, and sends it to the enabled integrations when it was not starred yet
- `as=unsaved`: unstars the entry

Notes:

- `saved` and `unsaved` set the starred state: sending them twice has the same effect as sending them once
- another `as` value changes nothing
- if `id` is missing, invalid or not positive, the server returns the base response without changes
- if the entry does not exist or belongs to another user, the server returns the base response without changes

### `mark=feed`

Parameters:

- `mark=feed`
- `as=read`
- `id=<feed_id>`
- `before=<unix_timestamp>`

Behavior:

- marks unread entries in the feed as read when `published_at < before`, before sending the response

Notes:

- if `id` is missing, invalid or not positive, the server returns the base response without changes
- if `before` is missing or invalid, it is treated as Unix time `0`, which usually means nothing is marked as read

### `mark=group`

Parameters:

- `mark=group`
- `as=read`
- `id=<group_id>`
- `before=<unix_timestamp>`

Behavior:

- `id=0`: marks all unread entries as read, or only those with `published_at < before` when `before` is a positive timestamp
- `id>0`: marks unread entries in the matching category as read when `published_at < before`
- the entries are marked before sending the response

Notes:

- group IDs map to Miniflux category IDs
- `id=0` must be sent explicitly: if `id` is missing, invalid or negative, the server returns the base response without changes
- if `before` is missing or invalid for `id>0`, it is treated as Unix time `0`, which usually means nothing is marked as read

## Error Handling

Authentication failures:

- HTTP status: `200`
- body: `{"api_version":3,"auth":0,"last_refreshed_on_time":0}`

Invalid write requests, such as a missing or invalid `id`, are not errors: they return the base authenticated response and change nothing.

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
- `mark=item` also accepts `as=unread`
- `mark=feed` and `mark=group` ignore the `as` parameter: they always mark entries as read

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
