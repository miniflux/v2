// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package database // import "miniflux.app/v2/internal/database"

import (
	"database/sql"

	"miniflux.app/v2/internal/crypto"
)

var schemaVersion = len(migrations)

// Order is important. Add new migrations at the end of the list.
var migrations = []func(tx *sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE schema_version (
				version text not null
			);

			CREATE TABLE users (
				id serial not null,
				username text not null unique,
				password text,
				is_admin bool default 'f',
				language text default 'en_US',
				timezone text default 'UTC',
				theme text default 'default',
				last_login_at timestamp with time zone,
				primary key (id)
			);

			CREATE TABLE sessions (
				id serial not null,
				user_id int not null,
				token text not null unique,
				created_at timestamp with time zone default now(),
				user_agent text,
				ip text,
				primary key (id),
				unique (user_id, token),
				foreign key (user_id) references users(id) on delete cascade
			);

			CREATE TABLE categories (
				id serial not null,
				user_id int not null,
				title text not null,
				primary key (id),
				unique (user_id, title),
				foreign key (user_id) references users(id) on delete cascade
			);

			CREATE TABLE feeds (
				id bigserial not null,
				user_id int not null,
				category_id int not null,
				title text not null,
				feed_url text not null,
				site_url text not null,
				checked_at timestamp with time zone default now(),
				etag_header text default '',
				last_modified_header text default '',
				parsing_error_msg text default '',
				parsing_error_count int default 0,
				primary key (id),
				unique (user_id, feed_url),
				foreign key (user_id) references users(id) on delete cascade,
				foreign key (category_id) references categories(id) on delete cascade
			);

			CREATE TYPE entry_status as enum('unread', 'read', 'removed');

			CREATE TABLE entries (
				id bigserial not null,
				user_id int not null,
				feed_id bigint not null,
				hash text not null,
				published_at timestamp with time zone not null,
				title text not null,
				url text not null,
				author text,
				content text,
				status entry_status default 'unread',
				primary key (id),
				unique (feed_id, hash),
				foreign key (user_id) references users(id) on delete cascade,
				foreign key (feed_id) references feeds(id) on delete cascade
			);

			CREATE INDEX entries_feed_idx on entries using btree(feed_id);

			CREATE TABLE enclosures (
				id bigserial not null,
				user_id int not null,
				entry_id bigint not null,
				url text not null,
				size int default 0,
				mime_type text default '',
				primary key (id),
				foreign key (user_id) references users(id) on delete cascade,
				foreign key (entry_id) references entries(id) on delete cascade
			);

			CREATE TABLE icons (
				id bigserial not null,
				hash text not null unique,
				mime_type text not null,
				content bytea not null,
				primary key (id)
			);

			CREATE TABLE feed_icons (
				feed_id bigint not null,
				icon_id bigint not null,
				primary key(feed_id, icon_id),
				foreign key (feed_id) references feeds(id) on delete cascade,
				foreign key (icon_id) references icons(id) on delete cascade
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE EXTENSION IF NOT EXISTS hstore;
			ALTER TABLE users ADD COLUMN extra hstore;
			CREATE INDEX users_extra_idx ON users using gin(extra);
			`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE tokens (
				id text not null,
				value text not null,
				created_at timestamp with time zone not null default now(),
				primary key(id, value)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TYPE entry_sorting_direction AS enum('asc', 'desc');
			ALTER TABLE users ADD COLUMN entry_direction entry_sorting_direction default 'asc';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE integrations (
				user_id int not null,
				pinboard_enabled bool default 'f',
				pinboard_token text default '',
				pinboard_tags text default 'miniflux',
				pinboard_mark_as_unread bool default 'f',
				instapaper_enabled bool default 'f',
				instapaper_username text default '',
				instapaper_password text default '',
				fever_enabled bool default 'f',
				fever_username text default '',
				fever_password text default '',
				fever_token text default '',
				primary key(user_id)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN scraper_rules text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN rewrite_rules text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN crawler boolean default 'f'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE sessions rename to user_sessions`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			DROP TABLE tokens;

			CREATE TABLE sessions (
				id text not null,
				data jsonb not null,
				created_at timestamp with time zone not null default now(),
				primary key(id)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN wallabag_enabled bool default 'f',
				ADD COLUMN wallabag_url text default '',
				ADD COLUMN wallabag_client_id text default '',
				ADD COLUMN wallabag_client_secret text default '',
				ADD COLUMN wallabag_username text default '',
				ADD COLUMN wallabag_password text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE entries ADD COLUMN starred bool default 'f'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE INDEX entries_user_status_idx ON entries(user_id, status);
			CREATE INDEX feeds_user_category_idx ON feeds(user_id, category_id);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN nunux_keeper_enabled bool default 'f',
				ADD COLUMN nunux_keeper_url text default '',
				ADD COLUMN nunux_keeper_api_key text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE enclosures ALTER COLUMN size SET DATA TYPE bigint`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE entries ADD COLUMN comments_url text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN pocket_enabled bool default 'f',
				ADD COLUMN pocket_access_token text default '',
				ADD COLUMN pocket_consumer_key text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE user_sessions ALTER COLUMN ip SET DATA TYPE inet using ip::inet;
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds
				ADD COLUMN username text default '',
				ADD COLUMN password text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE entries ADD COLUMN document_vectors tsvector;
			UPDATE entries SET document_vectors = to_tsvector(substring(title || ' ' || coalesce(content, '') for 1000000));
			CREATE INDEX document_vectors_idx ON entries USING gin(document_vectors);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN user_agent text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			UPDATE
				entries
			SET
				document_vectors = setweight(to_tsvector(substring(coalesce(title, '') for 1000000)), 'A') || setweight(to_tsvector(substring(coalesce(content, '') for 1000000)), 'B')
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN keyboard_shortcuts boolean default 't'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN disabled boolean default 'f';`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE users ALTER COLUMN theme SET DEFAULT 'light_serif';
			UPDATE users SET theme='light_serif' WHERE theme='default';
			UPDATE users SET theme='light_sans_serif' WHERE theme='sansserif';
			UPDATE users SET theme='dark_serif' WHERE theme='black';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE entries ADD COLUMN changed_at timestamp with time zone;
			UPDATE entries SET changed_at = published_at;
			ALTER TABLE entries ALTER COLUMN changed_at SET not null;
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE api_keys (
				id serial not null,
				user_id int not null references users(id) on delete cascade,
				token text not null unique,
				description text not null,
				last_used_at timestamp with time zone,
				created_at timestamp with time zone default now(),
				primary key(id),
				unique (user_id, description)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE entries ADD COLUMN share_code text not null default '';
			CREATE UNIQUE INDEX entries_share_code_idx ON entries USING btree(share_code) WHERE share_code <> '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `CREATE INDEX enclosures_user_entry_url_idx ON enclosures(user_id, entry_id, md5(url))`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds ADD COLUMN next_check_at timestamp with time zone default now();
			CREATE INDEX entries_user_feed_idx ON entries (user_id, feed_id);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN ignore_http_cache bool default false`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN entries_per_page int default 100`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN show_reading_time boolean default 't'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `CREATE INDEX entries_id_user_status_idx ON entries USING btree (id, user_id, status)`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN fetch_via_proxy bool default false`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `CREATE INDEX entries_feed_id_status_hash_idx ON entries USING btree (feed_id, status, hash)`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `CREATE INDEX entries_user_id_status_starred_idx ON entries (user_id, status, starred)`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN entry_swipe boolean default 't'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE integrations DROP COLUMN fever_password`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds
				ADD COLUMN blocklist_rules text not null default '',
				ADD COLUMN keeplist_rules text not null default ''
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE entries ADD COLUMN reading_time int not null default 0`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE entries ADD COLUMN created_at timestamp with time zone not null default now();
			UPDATE entries SET created_at = published_at;
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE users
				ADD column stylesheet text not null default '',
				ADD column google_id text not null default '',
				ADD column openid_connect_id text not null default ''
		`)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
				DECLARE my_cursor CURSOR FOR
				SELECT
					id,
					COALESCE(extra->'custom_css', '') as custom_css,
					COALESCE(extra->'google_id', '') as google_id,
					COALESCE(extra->'oidc_id', '') as oidc_id
				FROM users
				FOR UPDATE
			`)
		if err != nil {
			return err
		}
		defer tx.Exec("CLOSE my_cursor")

		for {
			var (
				userID           int64
				customStylesheet string
				googleID         string
				oidcID           string
			)

			if err := tx.QueryRow(`FETCH NEXT FROM my_cursor`).Scan(&userID, &customStylesheet, &googleID, &oidcID); err != nil {
				if err == sql.ErrNoRows {
					break
				}
				return err
			}

			_, err := tx.Exec(
				`UPDATE
						users
					SET
						stylesheet=$2,
						google_id=$3,
						openid_connect_id=$4
					WHERE
						id=$1
					`,
				userID, customStylesheet, googleID, oidcID)
			if err != nil {
				return err
			}
		}

		return err
	},
	func(tx *sql.Tx) (err error) {
		if _, err = tx.Exec(`ALTER TABLE users DROP COLUMN extra;`); err != nil {
			return err
		}
		_, err = tx.Exec(`
			CREATE UNIQUE INDEX users_google_id_idx ON users(google_id) WHERE google_id <> '';
			CREATE UNIQUE INDEX users_openid_connect_id_idx ON users(openid_connect_id) WHERE openid_connect_id <> '';
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE INDEX entries_feed_url_idx ON entries(feed_id, url) WHERE length(url) < 2000;
			CREATE INDEX entries_user_status_feed_idx ON entries(user_id, status, feed_id);
			CREATE INDEX entries_user_status_changed_idx ON entries(user_id, status, changed_at);
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE TABLE acme_cache (
				key varchar(400) not null primary key,
				data bytea not null,
				updated_at timestamptz not null
			);
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE feeds ADD COLUMN allow_self_signed_certificates boolean not null default false
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TYPE webapp_display_mode AS enum('fullscreen', 'standalone', 'minimal-ui', 'browser');
			ALTER TABLE users ADD COLUMN display_mode webapp_display_mode default 'standalone';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN cookie text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE categories ADD COLUMN hide_globally boolean not null default false
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE feeds ADD COLUMN hide_globally boolean not null default false
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN telegram_bot_enabled bool default 'f',
				ADD COLUMN telegram_bot_token text default '',
				ADD COLUMN telegram_bot_chat_id text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TYPE entry_sorting_order AS enum('published_at', 'created_at');
			ALTER TABLE users ADD COLUMN entry_order entry_sorting_order default 'published_at';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN googlereader_enabled bool default 'f',
				ADD COLUMN googlereader_username text default '',
				ADD COLUMN googlereader_password text default '';
			`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN espial_enabled bool default 'f',
				ADD COLUMN espial_url text default '',
				ADD COLUMN espial_api_key text default '',
				ADD COLUMN espial_tags text default 'miniflux';
			`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN linkding_enabled bool default 'f',
				ADD COLUMN linkding_url text default '',
				ADD COLUMN linkding_api_key text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE feeds ADD COLUMN url_rewrite_rules text not null default ''
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE users
				ADD COLUMN default_reading_speed int default 265,
				ADD COLUMN cjk_reading_speed int default 500;
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE users ADD COLUMN default_home_page text default 'unread';
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE integrations ADD COLUMN wallabag_only_url bool default 'f';
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE users ADD COLUMN categories_sorting_order text not null default 'unread_count';
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN matrix_bot_enabled bool default 'f',
				ADD COLUMN matrix_bot_user text default '',
				ADD COLUMN matrix_bot_password text default '',
				ADD COLUMN matrix_bot_url text default '',
				ADD COLUMN matrix_bot_chat_id text default '';
		`
		_, err = tx.Exec(sql)
		return
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN double_tap boolean default 't'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE entries ADD COLUMN tags text[] default '{}';
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE users RENAME double_tap TO gesture_nav;
			ALTER TABLE users
				ALTER COLUMN gesture_nav SET DATA TYPE text using case when gesture_nav = true then 'tap' when gesture_nav = false then 'none' end,
				ALTER COLUMN gesture_nav SET default 'tap';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations ADD COLUMN linkding_tags text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds ADD COLUMN no_media_player boolean default 'f';
			ALTER TABLE enclosures ADD COLUMN media_progression int default 0;
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations ADD COLUMN linkding_mark_as_unread bool default 'f';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		// Delete duplicated rows
		sql := `
			DELETE FROM enclosures a USING enclosures b
			WHERE a.id < b.id
				AND a.user_id = b.user_id
				AND a.entry_id = b.entry_id
				AND a.url = b.url;
		`
		_, err = tx.Exec(sql)
		if err != nil {
			return err
		}

		// Remove previous index
		_, err = tx.Exec(`DROP INDEX enclosures_user_entry_url_idx`)
		if err != nil {
			return err
		}

		// Create unique index
		_, err = tx.Exec(`CREATE UNIQUE INDEX enclosures_user_entry_url_unique_idx ON enclosures(user_id, entry_id, md5(url))`)
		if err != nil {
			return err
		}

		return nil
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN mark_read_on_view boolean default 't'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN notion_enabled bool default 'f',
				ADD COLUMN notion_token text default '',
				ADD COLUMN notion_page_id text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN readwise_enabled bool default 'f',
				ADD COLUMN readwise_api_key text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN apprise_enabled bool default 'f',
				ADD COLUMN apprise_url text default '',
				ADD COLUMN apprise_services_url text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN shiori_enabled bool default 'f',
				ADD COLUMN shiori_url text default '',
				ADD COLUMN shiori_username text default '',
				ADD COLUMN shiori_password text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN shaarli_enabled bool default 'f',
				ADD COLUMN shaarli_url text default '',
				ADD COLUMN shaarli_api_secret text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			ALTER TABLE feeds ADD COLUMN apprise_service_urls text default '';
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN webhook_enabled bool default 'f',
				ADD COLUMN webhook_url text default '',
				ADD COLUMN webhook_secret text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN telegram_bot_topic_id int,
				ADD COLUMN telegram_bot_disable_web_page_preview bool default 'f',
				ADD COLUMN telegram_bot_disable_notification bool default 'f';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations ADD COLUMN telegram_bot_disable_buttons bool default 'f';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			-- Speed up has_enclosure
			CREATE INDEX enclosures_entry_id_idx ON enclosures(entry_id);

			-- Speed up unread page
			CREATE INDEX entries_user_status_published_idx ON entries(user_id, status, published_at);
			CREATE INDEX entries_user_status_created_idx ON entries(user_id, status, created_at);
			CREATE INDEX feeds_feed_id_hide_globally_idx ON feeds(id, hide_globally);

			-- Speed up history page
			CREATE INDEX entries_user_status_changed_published_idx ON entries(user_id, status, changed_at, published_at);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN rssbridge_enabled bool default 'f',
				ADD COLUMN rssbridge_url text default '';
		`
		_, err = tx.Exec(sql)
		return
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE TABLE webauthn_credentials (
				handle bytea primary key,
				cred_id bytea unique not null,
				user_id int references users(id) on delete cascade not null,
				public_key bytea not null,
				attestation_type varchar(255) not null,
				aaguid bytea,
				sign_count bigint,
				clone_warning bool,
				name text,
				added_on timestamp with time zone default now(),
				last_seen_on timestamp with time zone default now()
			);
		`)
		return
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN omnivore_enabled bool default 'f',
				ADD COLUMN omnivore_api_key text default '',
				ADD COLUMN omnivore_url text default '';
		`
		_, err = tx.Exec(sql)
		return
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN linkace_enabled bool default 'f',
				ADD COLUMN linkace_url text default '',
				ADD COLUMN linkace_api_key text default '',
				ADD COLUMN linkace_tags text default '',
				ADD COLUMN linkace_is_private bool default 't',
				ADD COLUMN linkace_check_disabled bool default 't';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN linkwarden_enabled bool default 'f',
				ADD COLUMN linkwarden_url text default '',
				ADD COLUMN linkwarden_api_key text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN readeck_enabled bool default 'f',
				ADD COLUMN readeck_only_url bool default 'f',
				ADD COLUMN readeck_url text default '',
				ADD COLUMN readeck_api_key text default '',
				ADD COLUMN readeck_labels text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN disable_http2 bool default 'f'`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN media_playback_rate numeric default 1;`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		// the WHERE part speed-up the request a lot
		sql := `UPDATE entries SET tags = array_remove(tags, '') WHERE '' = ANY(tags);`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		// Entry URLs can exceeds btree maximum size
		// Checking entry existence is now using entries_feed_id_status_hash_idx index
		_, err = tx.Exec(`DROP INDEX entries_feed_url_idx`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN raindrop_enabled bool default 'f',
				ADD COLUMN raindrop_token text default '',
				ADD COLUMN raindrop_collection_id text default '',
				ADD COLUMN raindrop_tags text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE feeds ADD COLUMN description text default ''`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE users
				ADD COLUMN block_filter_entry_rules text not null default '',
				ADD COLUMN keep_filter_entry_rules text not null default ''
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN betula_url text default '',
				ADD COLUMN betula_token text default '',
				ADD COLUMN betula_enabled bool default 'f';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN ntfy_enabled bool default 'f',
				ADD COLUMN ntfy_url text default '',
				ADD COLUMN ntfy_topic text default '',
				ADD COLUMN ntfy_api_token text default '',
				ADD COLUMN ntfy_username text default '',
				ADD COLUMN ntfy_password text default '',
				ADD COLUMN ntfy_icon_url text default '';

			ALTER TABLE feeds
				ADD COLUMN ntfy_enabled bool default 'f',
				ADD COLUMN ntfy_priority int default '3';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN mark_read_on_media_player_completion bool default 'f';`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN custom_js text not null default '';`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE users ADD COLUMN external_font_hosts text not null default '';`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN cubox_enabled bool default 'f',
				ADD COLUMN cubox_api_link text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN discord_enabled bool default 'f',
				ADD COLUMN discord_webhook_link text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `ALTER TABLE integrations ADD COLUMN ntfy_internal_links bool default 'f';`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN slack_enabled bool default 'f',
				ADD COLUMN slack_webhook_link text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`ALTER TABLE feeds ADD COLUMN webhook_url text default '';`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN pushover_enabled bool default 'f',
				ADD COLUMN pushover_user text default '',
				ADD COLUMN pushover_token text default '',
				ADD COLUMN pushover_device text default '',
				ADD COLUMN pushover_prefix text default '';

			ALTER TABLE feeds
				ADD COLUMN pushover_enabled bool default 'f',
				ADD COLUMN pushover_priority int default '0';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds ADD COLUMN ntfy_topic text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE icons ADD COLUMN external_id text default '';
			CREATE UNIQUE INDEX icons_external_id_idx ON icons USING btree(external_id) WHERE external_id <> '';
		`
		_, err = tx.Exec(sql)

		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
				DECLARE id_cursor CURSOR FOR
				SELECT
					id
				FROM icons
				WHERE external_id = ''
				FOR UPDATE`)
		if err != nil {
			return err
		}
		defer tx.Exec("CLOSE id_cursor")

		for {
			var id int64

			if err := tx.QueryRow(`FETCH NEXT FROM id_cursor`).Scan(&id); err != nil {
				if err == sql.ErrNoRows {
					break
				}
				return err
			}

			_, err = tx.Exec(
				`
				UPDATE icons SET external_id = $1 WHERE id = $2
				`,
				crypto.GenerateRandomStringHex(20), id)

			if err != nil {
				return err
			}
		}

		return nil
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`ALTER TABLE feeds ADD COLUMN proxy_url text default ''`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations ADD COLUMN rssbridge_token text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`ALTER TABLE users ADD COLUMN always_open_external_links bool default 'f'`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				ADD COLUMN karakeep_enabled bool default 'f',
			 	ADD COLUMN karakeep_api_key text default '',
				ADD COLUMN karakeep_url text default '';
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`ALTER TABLE users ADD COLUMN open_external_links_in_new_tab bool default 't'`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE integrations
				DROP COLUMN pocket_enabled,
				DROP COLUMN pocket_access_token,
				DROP COLUMN pocket_consumer_key;
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		sql := `
			ALTER TABLE feeds
				ADD COLUMN block_filter_entry_rules text not null default '',
				ADD COLUMN keep_filter_entry_rules text not null default ''
		`
		_, err = tx.Exec(sql)
		return err
	},
}
