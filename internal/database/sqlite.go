// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package database // import "miniflux.app/v2/internal/database"

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var sqliteSchemaVersion = len(sqliteMigrations)

// Order is important. Add new migrations at the end of the list.
var sqliteMigrations = []Migration{
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			CREATE TABLE schema_version (
				version INTEGER NOT NULL
			);
		`)
		return err
	},
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL UNIQUE,
				password TEXT,
				is_admin INTEGER NOT NULL DEFAULT 0,
				language TEXT NOT NULL DEFAULT 'en_US',
				timezone TEXT NOT NULL DEFAULT 'UTC',
				theme TEXT NOT NULL DEFAULT 'light_serif',
				last_login_at DATETIME,
				entry_direction TEXT NOT NULL DEFAULT 'asc' CHECK (entry_direction IN ('asc','desc')),
				keyboard_shortcuts INTEGER NOT NULL DEFAULT 1,
				entries_per_page INTEGER NOT NULL DEFAULT 100,
				show_reading_time INTEGER NOT NULL DEFAULT 1,
				entry_swipe INTEGER NOT NULL DEFAULT 1,
				stylesheet TEXT NOT NULL DEFAULT '',
				google_id TEXT NOT NULL DEFAULT '',
				openid_connect_id TEXT NOT NULL DEFAULT '',
				display_mode TEXT NOT NULL DEFAULT 'standalone' CHECK (display_mode IN ('fullscreen','standalone','minimal-ui','browser')),
				entry_order TEXT NOT NULL DEFAULT 'published_at' CHECK (entry_order IN ('published_at','created_at')),
				default_reading_speed INTEGER NOT NULL DEFAULT 265,
				cjk_reading_speed INTEGER NOT NULL DEFAULT 500,
				default_home_page TEXT NOT NULL DEFAULT 'unread',
				categories_sorting_order TEXT NOT NULL DEFAULT 'unread_count',
				gesture_nav TEXT NOT NULL DEFAULT 'tap',
				mark_read_on_view INTEGER NOT NULL DEFAULT 1,
				media_playback_rate REAL NOT NULL DEFAULT 1.0,
				block_filter_entry_rules TEXT NOT NULL DEFAULT '',
				keep_filter_entry_rules TEXT NOT NULL DEFAULT '',
				mark_read_on_media_player_completion INTEGER NOT NULL DEFAULT 0,
				custom_js TEXT NOT NULL DEFAULT '',
				external_font_hosts TEXT NOT NULL DEFAULT '',
				always_open_external_links INTEGER NOT NULL DEFAULT 0,
				open_external_links_in_new_tab INTEGER NOT NULL DEFAULT 1
			);
			CREATE TABLE user_sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token TEXT NOT NULL UNIQUE,
				created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
				user_agent TEXT,
				ip TEXT,
				UNIQUE(user_id, token),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE TABLE categories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				hide_globally INTEGER NOT NULL DEFAULT 0,
				UNIQUE(user_id, title),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE TABLE feeds (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				category_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				feed_url TEXT NOT NULL,
				site_url TEXT NOT NULL,
				checked_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
				etag_header TEXT NOT NULL DEFAULT '',
				last_modified_header TEXT NOT NULL DEFAULT '',
				parsing_error_msg TEXT NOT NULL DEFAULT '',
				parsing_error_count INTEGER NOT NULL DEFAULT 0,
				scraper_rules TEXT NOT NULL DEFAULT '',
				rewrite_rules TEXT NOT NULL DEFAULT '',
				crawler INTEGER NOT NULL DEFAULT 0,
				username TEXT NOT NULL DEFAULT '',
				password TEXT NOT NULL DEFAULT '',
				user_agent TEXT NOT NULL DEFAULT '',
				disabled INTEGER NOT NULL DEFAULT 0,
				next_check_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
				ignore_http_cache INTEGER NOT NULL DEFAULT 0,
				fetch_via_proxy INTEGER NOT NULL DEFAULT 0,
				blocklist_rules TEXT NOT NULL DEFAULT '',
				keeplist_rules TEXT NOT NULL DEFAULT '',
				allow_self_signed_certificates INTEGER NOT NULL DEFAULT 0,
				cookie TEXT NOT NULL DEFAULT '',
				hide_globally INTEGER NOT NULL DEFAULT 0,
				url_rewrite_rules TEXT NOT NULL DEFAULT '',
				no_media_player INTEGER NOT NULL DEFAULT 0,
				apprise_service_urls TEXT NOT NULL DEFAULT '',
				disable_http2 INTEGER NOT NULL DEFAULT 0,
				description TEXT NOT NULL DEFAULT '',
				ntfy_enabled INTEGER NOT NULL DEFAULT 0,
				ntfy_priority INTEGER NOT NULL DEFAULT 3,
				webhook_url TEXT NOT NULL DEFAULT '',
				pushover_enabled INTEGER NOT NULL DEFAULT 0,
				pushover_priority INTEGER NOT NULL DEFAULT 0,
				ntfy_topic TEXT NOT NULL DEFAULT '',
				proxy_url TEXT NOT NULL DEFAULT '',
				block_filter_entry_rules TEXT NOT NULL DEFAULT '',
				keep_filter_entry_rules TEXT NOT NULL DEFAULT '',
				UNIQUE(user_id, feed_url),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
			);
			CREATE TABLE entries (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				feed_id INTEGER NOT NULL,
				hash TEXT NOT NULL,
				published_at DATETIME NOT NULL,
				title TEXT NOT NULL,
				url TEXT NOT NULL,
				author TEXT,
				content TEXT,
				status TEXT NOT NULL DEFAULT 'unread' CHECK (status IN ('unread','read','removed')),
				starred INTEGER NOT NULL DEFAULT 0,
				comments_url TEXT NOT NULL DEFAULT '',
				changed_at DATETIME NOT NULL,
				share_code TEXT NOT NULL DEFAULT '',
				reading_time INTEGER NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
				tags TEXT NOT NULL DEFAULT '',
				UNIQUE(feed_id, hash),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
			);
			CREATE TABLE enclosures (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				entry_id INTEGER NOT NULL,
				url TEXT NOT NULL,
				size INTEGER NOT NULL DEFAULT 0,
				mime_type TEXT NOT NULL DEFAULT '',
				media_progression INTEGER NOT NULL DEFAULT 0,
				UNIQUE(user_id, entry_id, url),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
			);
			CREATE TABLE icons (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				hash TEXT NOT NULL,
				mime_type TEXT NOT NULL,
				content BLOB NOT NULL,
				external_id TEXT NOT NULL DEFAULT '',
				UNIQUE(hash)
			);
			CREATE TABLE feed_icons (
				feed_id INTEGER NOT NULL,
				icon_id INTEGER NOT NULL,
				PRIMARY KEY (feed_id, icon_id),
				FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
				FOREIGN KEY (icon_id) REFERENCES icons(id) ON DELETE CASCADE
			);
			CREATE TABLE integrations (
				user_id INTEGER PRIMARY KEY,
				pinboard_enabled INTEGER NOT NULL DEFAULT 0,
				pinboard_token TEXT NOT NULL DEFAULT '',
				pinboard_tags TEXT NOT NULL DEFAULT 'miniflux',
				pinboard_mark_as_unread INTEGER NOT NULL DEFAULT 0,
				instapaper_enabled INTEGER NOT NULL DEFAULT 0,
				instapaper_username TEXT NOT NULL DEFAULT '',
				instapaper_password TEXT NOT NULL DEFAULT '',
				fever_enabled INTEGER NOT NULL DEFAULT 0,
				fever_username TEXT NOT NULL DEFAULT '',
				fever_token TEXT NOT NULL DEFAULT '',
				wallabag_enabled INTEGER NOT NULL DEFAULT 0,
				wallabag_url TEXT NOT NULL DEFAULT '',
				wallabag_client_id TEXT NOT NULL DEFAULT '',
				wallabag_client_secret TEXT NOT NULL DEFAULT '',
				wallabag_username TEXT NOT NULL DEFAULT '',
				wallabag_password TEXT NOT NULL DEFAULT '',
				nunux_keeper_enabled INTEGER NOT NULL DEFAULT 0,
				nunux_keeper_url TEXT NOT NULL DEFAULT '',
				nunux_keeper_api_key TEXT NOT NULL DEFAULT '',
				telegram_bot_enabled INTEGER NOT NULL DEFAULT 0,
				telegram_bot_token TEXT NOT NULL DEFAULT '',
				telegram_bot_chat_id TEXT NOT NULL DEFAULT '',
				googlereader_enabled INTEGER NOT NULL DEFAULT 0,
				googlereader_username TEXT NOT NULL DEFAULT '',
				googlereader_password TEXT NOT NULL DEFAULT '',
				espial_enabled INTEGER NOT NULL DEFAULT 0,
				espial_url TEXT NOT NULL DEFAULT '',
				espial_api_key TEXT NOT NULL DEFAULT '',
				espial_tags TEXT NOT NULL DEFAULT 'miniflux',
				linkding_enabled INTEGER NOT NULL DEFAULT 0,
				linkding_url TEXT NOT NULL DEFAULT '',
				linkding_api_key TEXT NOT NULL DEFAULT '',
				wallabag_only_url INTEGER NOT NULL DEFAULT 0,
				matrix_bot_enabled INTEGER NOT NULL DEFAULT 0,
				matrix_bot_user TEXT NOT NULL DEFAULT '',
				matrix_bot_password TEXT NOT NULL DEFAULT '',
				matrix_bot_url TEXT NOT NULL DEFAULT '',
				matrix_bot_chat_id TEXT NOT NULL DEFAULT '',
				linkding_tags TEXT NOT NULL DEFAULT '',
				linkding_mark_as_unread INTEGER NOT NULL DEFAULT 0,
				notion_enabled INTEGER NOT NULL DEFAULT 0,
				notion_token TEXT NOT NULL DEFAULT '',
				notion_page_id TEXT NOT NULL DEFAULT '',
				readwise_enabled INTEGER NOT NULL DEFAULT 0,
				readwise_api_key TEXT NOT NULL DEFAULT '',
				apprise_enabled INTEGER NOT NULL DEFAULT 0,
				apprise_url TEXT NOT NULL DEFAULT '',
				apprise_services_url TEXT NOT NULL DEFAULT '',
				shiori_enabled INTEGER NOT NULL DEFAULT 0,
				shiori_url TEXT NOT NULL DEFAULT '',
				shiori_username TEXT NOT NULL DEFAULT '',
				shiori_password TEXT NOT NULL DEFAULT '',
				shaarli_enabled INTEGER NOT NULL DEFAULT 0,
				shaarli_url TEXT NOT NULL DEFAULT '',
				shaarli_api_secret TEXT NOT NULL DEFAULT '',
				webhook_enabled INTEGER NOT NULL DEFAULT 0,
				webhook_url TEXT NOT NULL DEFAULT '',
				webhook_secret TEXT NOT NULL DEFAULT '',
				telegram_bot_topic_id INTEGER,
				telegram_bot_disable_web_page_preview INTEGER NOT NULL DEFAULT 0,
				telegram_bot_disable_notification INTEGER NOT NULL DEFAULT 0,
				telegram_bot_disable_buttons INTEGER NOT NULL DEFAULT 0,
				rssbridge_enabled INTEGER NOT NULL DEFAULT 0,
				rssbridge_url TEXT NOT NULL DEFAULT '',
				omnivore_enabled INTEGER NOT NULL DEFAULT 0,
				omnivore_api_key TEXT NOT NULL DEFAULT '',
				omnivore_url TEXT NOT NULL DEFAULT '',
				linkace_enabled INTEGER NOT NULL DEFAULT 0,
				linkace_url TEXT NOT NULL DEFAULT '',
				linkace_api_key TEXT NOT NULL DEFAULT '',
				linkace_tags TEXT NOT NULL DEFAULT '',
				linkace_is_private INTEGER NOT NULL DEFAULT 1,
				linkace_check_disabled INTEGER NOT NULL DEFAULT 1,
				linkwarden_enabled INTEGER NOT NULL DEFAULT 0,
				linkwarden_url TEXT NOT NULL DEFAULT '',
				linkwarden_api_key TEXT NOT NULL DEFAULT '',
				readeck_enabled INTEGER NOT NULL DEFAULT 0,
				readeck_only_url INTEGER NOT NULL DEFAULT 0,
				readeck_url TEXT NOT NULL DEFAULT '',
				readeck_api_key TEXT NOT NULL DEFAULT '',
				readeck_labels TEXT NOT NULL DEFAULT '',
				raindrop_enabled INTEGER NOT NULL DEFAULT 0,
				raindrop_token TEXT NOT NULL DEFAULT '',
				raindrop_collection_id TEXT NOT NULL DEFAULT '',
				raindrop_tags TEXT NOT NULL DEFAULT '',
				betula_url TEXT NOT NULL DEFAULT '',
				betula_token TEXT NOT NULL DEFAULT '',
				betula_enabled INTEGER NOT NULL DEFAULT 0,
				ntfy_enabled INTEGER NOT NULL DEFAULT 0,
				ntfy_url TEXT NOT NULL DEFAULT '',
				ntfy_topic TEXT NOT NULL DEFAULT '',
				ntfy_api_token TEXT NOT NULL DEFAULT '',
				ntfy_username TEXT NOT NULL DEFAULT '',
				ntfy_password TEXT NOT NULL DEFAULT '',
				ntfy_icon_url TEXT NOT NULL DEFAULT '',
				cubox_enabled INTEGER NOT NULL DEFAULT 0,
				cubox_api_link TEXT NOT NULL DEFAULT '',
				discord_enabled INTEGER NOT NULL DEFAULT 0,
				discord_webhook_link TEXT NOT NULL DEFAULT '',
				ntfy_internal_links INTEGER NOT NULL DEFAULT 0,
				slack_enabled INTEGER NOT NULL DEFAULT 0,
				slack_webhook_link TEXT NOT NULL DEFAULT '',
				pushover_enabled INTEGER NOT NULL DEFAULT 0,
				pushover_user TEXT NOT NULL DEFAULT '',
				pushover_token TEXT NOT NULL DEFAULT '',
				pushover_device TEXT NOT NULL DEFAULT '',
				pushover_prefix TEXT NOT NULL DEFAULT '',
				rssbridge_token TEXT NOT NULL DEFAULT '',
				karakeep_enabled INTEGER NOT NULL DEFAULT 0,
				karakeep_api_key TEXT NOT NULL DEFAULT '',
				karakeep_url TEXT NOT NULL DEFAULT '',
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE TABLE sessions (
				id TEXT PRIMARY KEY,
				data TEXT NOT NULL,
				created_at DATETIME NOT NULL DEFAULT (DATETIME('now'))
			);
			CREATE TABLE api_keys (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token TEXT NOT NULL UNIQUE,
				description TEXT NOT NULL,
				last_used_at DATETIME,
				created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
				UNIQUE(user_id, description),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE TABLE acme_cache (
				key TEXT PRIMARY KEY,
				data BLOB NOT NULL,
				updated_at DATETIME NOT NULL
			);
			CREATE TABLE webauthn_credentials (
				handle BLOB PRIMARY KEY,
				cred_id BLOB NOT NULL UNIQUE,
				user_id INTEGER NOT NULL,
				key BLOB NOT NULL,
				attestation_type TEXT NOT NULL,
				aaguid BLOB,
				sign_count INTEGER,
				clone_warning INTEGER,
				name TEXT,
				added_on DATETIME NOT NULL DEFAULT (DATETIME('now')),
				last_seen_on DATETIME NOT NULL DEFAULT (DATETIME('now')),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
		`)
		return err
	},
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			CREATE UNIQUE INDEX icons_external_id_uq ON icons(external_id) WHERE external_id != '';
			CREATE UNIQUE INDEX users_google_id_uq ON users(google_id) WHERE google_id != '';
			CREATE UNIQUE INDEX users_openid_connect_id_uq ON users(openid_connect_id) WHERE openid_connect_id != '';
			CREATE UNIQUE INDEX entries_share_code_uq ON entries(share_code) WHERE share_code != '';

			CREATE INDEX feeds_user_category_idx ON feeds(user_id, category_id);
			CREATE INDEX feeds_id_hide_globally_idx ON feeds(id, hide_globally);
			CREATE INDEX entries_feed_idx ON entries(feed_id);
			CREATE INDEX entries_user_status_idx ON entries(user_id, status);
			CREATE INDEX entries_user_feed_idx ON entries(user_id, feed_id);
			CREATE INDEX entries_user_status_changed_idx ON entries(user_id, status, changed_at);
			CREATE INDEX entries_user_status_published_idx ON entries(user_id, status, published_at);
			CREATE INDEX entries_user_status_created_idx ON entries(user_id, status, created_at);
			CREATE INDEX entries_id_user_status_idx ON entries(id, user_id, status);
			CREATE INDEX entries_feed_id_status_hash_idx ON entries(feed_id, status, hash);
			CREATE INDEX entries_user_id_status_starred_idx ON entries(user_id, status, starred);
			CREATE INDEX entries_user_status_feed_idx ON entries(user_id, status, feed_id);
			CREATE INDEX entries_user_status_changed_published_idx ON entries(user_id, status, changed_at, published_at);
			CREATE INDEX enclosures_entry_id_idx ON enclosures(entry_id);
			CREATE INDEX feed_icons_icon_id_idx ON feed_icons(icon_id);
		`)
		return err
	},
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			CREATE VIRTUAL TABLE entries_fts USING fts5(
				title,
				content,
				tags,
				entry_id UNINDEXED,
				content='',
				-- as close to PostgreSQL as possible while being multilingual
    			tokenize = "unicode61 remove_diacritics 2 tokenchars '-_'"
  			);

			CREATE TRIGGER entries_ai AFTER INSERT ON entries BEGIN
				INSERT INTO entries_fts(rowid,title,content,tags,entry_id)
				VALUES (new.id,new.title,COALESCE(new.content,''),COALESCE(new.tags,''),new.id);
			END;

			CREATE TRIGGER entries_au AFTER UPDATE ON entries BEGIN
				INSERT INTO entries_fts(entries_fts,rowid) VALUES('delete', old.id);
				INSERT INTO entries_fts(rowid,title,content,tags,entry_id)
				VALUES (new.id,new.title,COALESCE(new.content,''),COALESCE(new.tags,''),new.id);
			END;

			CREATE TRIGGER entries_ad AFTER DELETE ON entries BEGIN
				INSERT INTO entries_fts(entries_fts,rowid) VALUES('delete', old.id);
			END;
		`)
		return err
	},
}
