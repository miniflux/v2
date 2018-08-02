create table integrations (
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
)
