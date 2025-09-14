// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package database // import "miniflux.app/v2/internal/database"

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var cockroachSchemaVersion = len(cockroachMigrations)

// Order is important. Add new migrations at the end of the list.
var cockroachMigrations = []Migration{
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE schema_version (
			  version STRING NOT NULL,
			  rowid INT8 NOT VISIBLE NOT NULL DEFAULT unique_rowid(),
			  CONSTRAINT schema_version_pkey PRIMARY KEY (rowid ASC)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TYPE entry_status AS ENUM ('unread', 'read', 'removed');
			CREATE TYPE entry_sorting_direction AS ENUM ('asc', 'desc');
			CREATE TYPE webapp_display_mode AS ENUM ('fullscreen', 'standalone', 'minimal-ui', 'browser');
			CREATE TYPE entry_sorting_order AS ENUM ('published_at', 'created_at');
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE users (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  username STRING NOT NULL,
			  password STRING NULL,
			  is_admin BOOL NULL DEFAULT false,
			  language STRING NULL DEFAULT 'en_US':::STRING,
			  timezone STRING NULL DEFAULT 'UTC':::STRING,
			  theme STRING NULL DEFAULT 'light_serif':::STRING,
			  last_login_at TIMESTAMPTZ NULL,
			  entry_direction entry_sorting_direction NULL DEFAULT 'asc':::entry_sorting_direction,
			  keyboard_shortcuts BOOL NULL DEFAULT true,
			  entries_per_page INT8 NULL DEFAULT 100:::INT8,
			  show_reading_time BOOL NULL DEFAULT true,
			  entry_swipe BOOL NULL DEFAULT true,
			  stylesheet STRING NOT NULL DEFAULT '':::STRING,
			  google_id STRING NOT NULL DEFAULT '':::STRING,
			  openid_connect_id STRING NOT NULL DEFAULT '':::STRING,
			  display_mode webapp_display_mode NULL DEFAULT 'standalone':::webapp_display_mode,
			  entry_order entry_sorting_order NULL DEFAULT 'published_at':::entry_sorting_order,
			  default_reading_speed INT8 NULL DEFAULT 265:::INT8,
			  cjk_reading_speed INT8 NULL DEFAULT 500:::INT8,
			  default_home_page STRING NULL DEFAULT 'unread':::STRING,
			  categories_sorting_order STRING NOT NULL DEFAULT 'unread_count':::STRING,
			  gesture_nav STRING NOT NULL DEFAULT 'tap':::STRING,
			  mark_read_on_view BOOL NULL DEFAULT true,
			  media_playback_rate DECIMAL NULL DEFAULT 1:::DECIMAL,
			  block_filter_entry_rules STRING NOT NULL DEFAULT '':::STRING,
			  keep_filter_entry_rules STRING NOT NULL DEFAULT '':::STRING,
			  mark_read_on_media_player_completion BOOL NULL DEFAULT false,
			  custom_js STRING NOT NULL DEFAULT '':::STRING,
			  external_font_hosts STRING NOT NULL DEFAULT '':::STRING,
			  always_open_external_links BOOL NULL DEFAULT false,
			  open_external_links_in_new_tab BOOL NULL DEFAULT true,
			  CONSTRAINT users_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX users_username_key (username ASC),
			  UNIQUE INDEX users_google_id_idx (google_id ASC) WHERE google_id != '':::STRING,
			  UNIQUE INDEX users_openid_connect_id_idx (openid_connect_id ASC) WHERE openid_connect_id != '':::STRING
			);
			CREATE TABLE user_sessions (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  token STRING NOT NULL,
			  created_at TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  user_agent STRING NULL,
			  ip STRING NULL,
			  CONSTRAINT sessions_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX sessions_token_key (token ASC),
			  UNIQUE INDEX sessions_user_id_token_key (user_id ASC, token ASC)
			);
			CREATE TABLE categories (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  title STRING NOT NULL,
			  hide_globally BOOL NOT NULL DEFAULT false,
			  CONSTRAINT categories_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX categories_user_id_title_key (user_id ASC, title ASC)
			);
			CREATE TABLE feeds (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  category_id INT8 NOT NULL,
			  title STRING NOT NULL,
			  feed_url STRING NOT NULL,
			  site_url STRING NOT NULL,
			  checked_at TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  etag_header STRING NULL DEFAULT '':::STRING,
			  last_modified_header STRING NULL DEFAULT '':::STRING,
			  parsing_error_msg STRING NULL DEFAULT '':::STRING,
			  parsing_error_count INT8 NULL DEFAULT 0:::INT8,
			  scraper_rules STRING NULL DEFAULT '':::STRING,
			  rewrite_rules STRING NULL DEFAULT '':::STRING,
			  crawler BOOL NULL DEFAULT false,
			  username STRING NULL DEFAULT '':::STRING,
			  password STRING NULL DEFAULT '':::STRING,
			  user_agent STRING NULL DEFAULT '':::STRING,
			  disabled BOOL NULL DEFAULT false,
			  next_check_at TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  ignore_http_cache BOOL NULL DEFAULT false,
			  fetch_via_proxy BOOL NULL DEFAULT false,
			  blocklist_rules STRING NOT NULL DEFAULT '':::STRING,
			  keeplist_rules STRING NOT NULL DEFAULT '':::STRING,
			  allow_self_signed_certificates BOOL NOT NULL DEFAULT false,
			  cookie STRING NULL DEFAULT '':::STRING,
			  hide_globally BOOL NOT NULL DEFAULT false,
			  url_rewrite_rules STRING NOT NULL DEFAULT '':::STRING,
			  no_media_player BOOL NULL DEFAULT false,
			  apprise_service_urls STRING NULL DEFAULT '':::STRING,
			  disable_http2 BOOL NULL DEFAULT false,
			  description STRING NULL DEFAULT '':::STRING,
			  ntfy_enabled BOOL NULL DEFAULT false,
			  ntfy_priority INT8 NULL DEFAULT 3:::INT8,
			  webhook_url STRING NULL DEFAULT '':::STRING,
			  pushover_enabled BOOL NULL DEFAULT false,
			  pushover_priority INT8 NULL DEFAULT 0:::INT8,
			  ntfy_topic STRING NULL DEFAULT '':::STRING,
			  proxy_url STRING NULL DEFAULT '':::STRING,
			  block_filter_entry_rules STRING NOT NULL DEFAULT '':::STRING,
			  keep_filter_entry_rules STRING NOT NULL DEFAULT '':::STRING,
			  CONSTRAINT feeds_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX feeds_user_id_feed_url_key (user_id ASC, feed_url ASC),
			  INDEX feeds_user_category_idx (user_id ASC, category_id ASC),
			  INDEX feeds_feed_id_hide_globally_idx (id ASC, hide_globally ASC)
			);
			CREATE TABLE entries (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  feed_id INT8 NOT NULL,
			  hash STRING NOT NULL,
			  published_at TIMESTAMPTZ NOT NULL,
			  title STRING NOT NULL,
			  url STRING NOT NULL,
			  author STRING NULL,
			  content STRING NULL,
			  status entry_status NULL DEFAULT 'unread':::entry_status,
			  starred BOOL NULL DEFAULT false,
			  comments_url STRING NULL DEFAULT '':::STRING,
			  document_vectors TSVECTOR NULL,
			  changed_at TIMESTAMPTZ NOT NULL,
			  share_code STRING NOT NULL DEFAULT '':::STRING,
			  reading_time INT8 NOT NULL DEFAULT 0:::INT8,
			  created_at TIMESTAMPTZ NOT NULL DEFAULT now():::TIMESTAMPTZ,
			  tags STRING[] NULL DEFAULT ARRAY[]:::STRING[],
			  CONSTRAINT entries_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX entries_feed_id_hash_key (feed_id ASC, hash ASC),
			  INDEX entries_feed_idx (feed_id ASC),
			  INDEX entries_user_status_idx (user_id ASC, status ASC),
			  INVERTED INDEX document_vectors_idx (document_vectors),
			  UNIQUE INDEX entries_share_code_idx (share_code ASC) WHERE share_code != '':::STRING,
			  INDEX entries_user_feed_idx (user_id ASC, feed_id ASC),
			  INDEX entries_id_user_status_idx (id ASC, user_id ASC, status ASC),
			  INDEX entries_feed_id_status_hash_idx (feed_id ASC, status ASC, hash ASC),
			  INDEX entries_user_id_status_starred_idx (user_id ASC, status ASC, starred ASC),
			  INDEX entries_user_status_feed_idx (user_id ASC, status ASC, feed_id ASC),
			  INDEX entries_user_status_changed_idx (user_id ASC, status ASC, changed_at ASC),
			  INDEX entries_user_status_published_idx (user_id ASC, status ASC, published_at ASC),
			  INDEX entries_user_status_created_idx (user_id ASC, status ASC, created_at ASC),
			  INDEX entries_user_status_changed_published_idx (user_id ASC, status ASC, changed_at ASC, published_at ASC)
			);
			CREATE TABLE enclosures (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  entry_id INT8 NOT NULL,
			  url STRING NOT NULL,
			  size INT8 NULL DEFAULT 0:::INT8,
			  mime_type STRING NULL DEFAULT '':::STRING,
			  media_progression INT8 NULL DEFAULT 0:::INT8,
			  CONSTRAINT enclosures_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX enclosures_user_entry_url_unique_idx (user_id ASC, entry_id ASC, url ASC),
			  INDEX enclosures_entry_id_idx (entry_id ASC)
			);
			CREATE TABLE icons (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  hash STRING NOT NULL,
			  mime_type STRING NOT NULL,
			  content BYTES NOT NULL,
			  external_id STRING NULL DEFAULT '':::STRING,
			  CONSTRAINT icons_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX icons_hash_key (hash ASC),
			  UNIQUE INDEX icons_external_id_idx (external_id ASC) WHERE external_id != '':::STRING
			);
			CREATE TABLE feed_icons (
			  feed_id INT8 NOT NULL,
			  icon_id INT8 NOT NULL,
			  CONSTRAINT feed_icons_pkey PRIMARY KEY (feed_id ASC, icon_id ASC)
			);
			CREATE TABLE integrations (
			  user_id INT8 NOT NULL,
			  pinboard_enabled BOOL NULL DEFAULT false,
			  pinboard_token STRING NULL DEFAULT '':::STRING,
			  pinboard_tags STRING NULL DEFAULT 'miniflux':::STRING,
			  pinboard_mark_as_unread BOOL NULL DEFAULT false,
			  instapaper_enabled BOOL NULL DEFAULT false,
			  instapaper_username STRING NULL DEFAULT '':::STRING,
			  instapaper_password STRING NULL DEFAULT '':::STRING,
			  fever_enabled BOOL NULL DEFAULT false,
			  fever_username STRING NULL DEFAULT '':::STRING,
			  fever_token STRING NULL DEFAULT '':::STRING,
			  wallabag_enabled BOOL NULL DEFAULT false,
			  wallabag_url STRING NULL DEFAULT '':::STRING,
			  wallabag_client_id STRING NULL DEFAULT '':::STRING,
			  wallabag_client_secret STRING NULL DEFAULT '':::STRING,
			  wallabag_username STRING NULL DEFAULT '':::STRING,
			  wallabag_password STRING NULL DEFAULT '':::STRING,
			  nunux_keeper_enabled BOOL NULL DEFAULT false,
			  nunux_keeper_url STRING NULL DEFAULT '':::STRING,
			  nunux_keeper_api_key STRING NULL DEFAULT '':::STRING,
			  telegram_bot_enabled BOOL NULL DEFAULT false,
			  telegram_bot_token STRING NULL DEFAULT '':::STRING,
			  telegram_bot_chat_id STRING NULL DEFAULT '':::STRING,
			  googlereader_enabled BOOL NULL DEFAULT false,
			  googlereader_username STRING NULL DEFAULT '':::STRING,
			  googlereader_password STRING NULL DEFAULT '':::STRING,
			  espial_enabled BOOL NULL DEFAULT false,
			  espial_url STRING NULL DEFAULT '':::STRING,
			  espial_api_key STRING NULL DEFAULT '':::STRING,
			  espial_tags STRING NULL DEFAULT 'miniflux':::STRING,
			  linkding_enabled BOOL NULL DEFAULT false,
			  linkding_url STRING NULL DEFAULT '':::STRING,
			  linkding_api_key STRING NULL DEFAULT '':::STRING,
			  wallabag_only_url BOOL NULL DEFAULT false,
			  matrix_bot_enabled BOOL NULL DEFAULT false,
			  matrix_bot_user STRING NULL DEFAULT '':::STRING,
			  matrix_bot_password STRING NULL DEFAULT '':::STRING,
			  matrix_bot_url STRING NULL DEFAULT '':::STRING,
			  matrix_bot_chat_id STRING NULL DEFAULT '':::STRING,
			  linkding_tags STRING NULL DEFAULT '':::STRING,
			  linkding_mark_as_unread BOOL NULL DEFAULT false,
			  notion_enabled BOOL NULL DEFAULT false,
			  notion_token STRING NULL DEFAULT '':::STRING,
			  notion_page_id STRING NULL DEFAULT '':::STRING,
			  readwise_enabled BOOL NULL DEFAULT false,
			  readwise_api_key STRING NULL DEFAULT '':::STRING,
			  apprise_enabled BOOL NULL DEFAULT false,
			  apprise_url STRING NULL DEFAULT '':::STRING,
			  apprise_services_url STRING NULL DEFAULT '':::STRING,
			  shiori_enabled BOOL NULL DEFAULT false,
			  shiori_url STRING NULL DEFAULT '':::STRING,
			  shiori_username STRING NULL DEFAULT '':::STRING,
			  shiori_password STRING NULL DEFAULT '':::STRING,
			  shaarli_enabled BOOL NULL DEFAULT false,
			  shaarli_url STRING NULL DEFAULT '':::STRING,
			  shaarli_api_secret STRING NULL DEFAULT '':::STRING,
			  webhook_enabled BOOL NULL DEFAULT false,
			  webhook_url STRING NULL DEFAULT '':::STRING,
			  webhook_secret STRING NULL DEFAULT '':::STRING,
			  telegram_bot_topic_id INT8 NULL,
			  telegram_bot_disable_web_page_preview BOOL NULL DEFAULT false,
			  telegram_bot_disable_notification BOOL NULL DEFAULT false,
			  telegram_bot_disable_buttons BOOL NULL DEFAULT false,
			  rssbridge_enabled BOOL NULL DEFAULT false,
			  rssbridge_url STRING NULL DEFAULT '':::STRING,
			  omnivore_enabled BOOL NULL DEFAULT false,
			  omnivore_api_key STRING NULL DEFAULT '':::STRING,
			  omnivore_url STRING NULL DEFAULT '':::STRING,
			  linkace_enabled BOOL NULL DEFAULT false,
			  linkace_url STRING NULL DEFAULT '':::STRING,
			  linkace_api_key STRING NULL DEFAULT '':::STRING,
			  linkace_tags STRING NULL DEFAULT '':::STRING,
			  linkace_is_private BOOL NULL DEFAULT true,
			  linkace_check_disabled BOOL NULL DEFAULT true,
			  linkwarden_enabled BOOL NULL DEFAULT false,
			  linkwarden_url STRING NULL DEFAULT '':::STRING,
			  linkwarden_api_key STRING NULL DEFAULT '':::STRING,
			  readeck_enabled BOOL NULL DEFAULT false,
			  readeck_only_url BOOL NULL DEFAULT false,
			  readeck_url STRING NULL DEFAULT '':::STRING,
			  readeck_api_key STRING NULL DEFAULT '':::STRING,
			  readeck_labels STRING NULL DEFAULT '':::STRING,
			  raindrop_enabled BOOL NULL DEFAULT false,
			  raindrop_token STRING NULL DEFAULT '':::STRING,
			  raindrop_collection_id STRING NULL DEFAULT '':::STRING,
			  raindrop_tags STRING NULL DEFAULT '':::STRING,
			  betula_url STRING NULL DEFAULT '':::STRING,
			  betula_token STRING NULL DEFAULT '':::STRING,
			  betula_enabled BOOL NULL DEFAULT false,
			  ntfy_enabled BOOL NULL DEFAULT false,
			  ntfy_url STRING NULL DEFAULT '':::STRING,
			  ntfy_topic STRING NULL DEFAULT '':::STRING,
			  ntfy_api_token STRING NULL DEFAULT '':::STRING,
			  ntfy_username STRING NULL DEFAULT '':::STRING,
			  ntfy_password STRING NULL DEFAULT '':::STRING,
			  ntfy_icon_url STRING NULL DEFAULT '':::STRING,
			  cubox_enabled BOOL NULL DEFAULT false,
			  cubox_api_link STRING NULL DEFAULT '':::STRING,
			  discord_enabled BOOL NULL DEFAULT false,
			  discord_webhook_link STRING NULL DEFAULT '':::STRING,
			  ntfy_internal_links BOOL NULL DEFAULT false,
			  slack_enabled BOOL NULL DEFAULT false,
			  slack_webhook_link STRING NULL DEFAULT '':::STRING,
			  pushover_enabled BOOL NULL DEFAULT false,
			  pushover_user STRING NULL DEFAULT '':::STRING,
			  pushover_token STRING NULL DEFAULT '':::STRING,
			  pushover_device STRING NULL DEFAULT '':::STRING,
			  pushover_prefix STRING NULL DEFAULT '':::STRING,
			  rssbridge_token STRING NULL DEFAULT '':::STRING,
			  karakeep_enabled BOOL NULL DEFAULT false,
			  karakeep_api_key STRING NULL DEFAULT '':::STRING,
			  karakeep_url STRING NULL DEFAULT '':::STRING,
			  CONSTRAINT integrations_pkey PRIMARY KEY (user_id ASC)
			);
			CREATE TABLE sessions (
			  id STRING NOT NULL,
			  data JSONB NOT NULL,
			  created_at TIMESTAMPTZ NOT NULL DEFAULT now():::TIMESTAMPTZ,
			  CONSTRAINT sessions_pkey PRIMARY KEY (id ASC)
			);
			CREATE TABLE api_keys (
			  id INT8 NOT NULL DEFAULT unique_rowid(),
			  user_id INT8 NOT NULL,
			  token STRING NOT NULL,
			  description STRING NOT NULL,
			  last_used_at TIMESTAMPTZ NULL,
			  created_at TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  CONSTRAINT api_keys_pkey PRIMARY KEY (id ASC),
			  UNIQUE INDEX api_keys_token_key (token ASC),
			  UNIQUE INDEX api_keys_user_id_description_key (user_id ASC, description ASC)
			);
			CREATE TABLE acme_cache (
			  key VARCHAR(400) NOT NULL,
			  data BYTES NOT NULL,
			  updated_at TIMESTAMPTZ NOT NULL,
			  CONSTRAINT acme_cache_pkey PRIMARY KEY (key ASC)
			);
			CREATE TABLE webauthn_credentials (
			  handle BYTES NOT NULL,
			  cred_id BYTES NOT NULL,
			  user_id INT8 NOT NULL,
			  key BYTES NOT NULL,
			  attestation_type VARCHAR(255) NOT NULL,
			  aaguid BYTES NULL,
			  sign_count INT8 NULL,
			  clone_warning BOOL NULL,
			  name STRING NULL,
			  added_on TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  last_seen_on TIMESTAMPTZ NULL DEFAULT now():::TIMESTAMPTZ,
			  CONSTRAINT webauthn_credentials_pkey PRIMARY KEY (handle ASC),
			  UNIQUE INDEX webauthn_credentials_cred_id_key (cred_id ASC)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE user_sessions ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE categories ADD CONSTRAINT categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE feeds ADD CONSTRAINT feeds_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE feeds ADD CONSTRAINT feeds_category_id_fkey FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;
			ALTER TABLE entries ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE entries ADD CONSTRAINT entries_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE;
			ALTER TABLE enclosures ADD CONSTRAINT enclosures_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE enclosures ADD CONSTRAINT enclosures_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE;
			ALTER TABLE feed_icons ADD CONSTRAINT feed_icons_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE;
			ALTER TABLE feed_icons ADD CONSTRAINT feed_icons_icon_id_fkey FOREIGN KEY (icon_id) REFERENCES icons(id) ON DELETE CASCADE;
			ALTER TABLE api_keys ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			ALTER TABLE webauthn_credentials ADD CONSTRAINT webauthn_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE user_sessions VALIDATE CONSTRAINT sessions_user_id_fkey;
			ALTER TABLE categories VALIDATE CONSTRAINT categories_user_id_fkey;
			ALTER TABLE feeds VALIDATE CONSTRAINT feeds_user_id_fkey;
			ALTER TABLE feeds VALIDATE CONSTRAINT feeds_category_id_fkey;
			ALTER TABLE entries VALIDATE CONSTRAINT entries_user_id_fkey;
			ALTER TABLE entries VALIDATE CONSTRAINT entries_feed_id_fkey;
			ALTER TABLE enclosures VALIDATE CONSTRAINT enclosures_user_id_fkey;
			ALTER TABLE enclosures VALIDATE CONSTRAINT enclosures_entry_id_fkey;
			ALTER TABLE feed_icons VALIDATE CONSTRAINT feed_icons_feed_id_fkey;
			ALTER TABLE feed_icons VALIDATE CONSTRAINT feed_icons_icon_id_fkey;
			ALTER TABLE api_keys VALIDATE CONSTRAINT api_keys_user_id_fkey;
			ALTER TABLE webauthn_credentials VALIDATE CONSTRAINT webauthn_credentials_user_id_fkey;
		`
		_, err = tx.Exec(sql)
		return err
	},
}
