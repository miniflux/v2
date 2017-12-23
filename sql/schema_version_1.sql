create table schema_version (
    version text not null
);

create table users (
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

create table sessions (
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

create table categories (
    id serial not null,
    user_id int not null,
    title text not null,
    primary key (id),
    unique (user_id, title),
    foreign key (user_id) references users(id) on delete cascade
);

create table feeds (
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

create type entry_status as enum('unread', 'read', 'removed');

create table entries (
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

create index entries_feed_idx on entries using btree(feed_id);

create table enclosures (
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

create table icons (
    id bigserial not null,
    hash text not null unique,
    mime_type text not null,
    content bytea not null,
    primary key (id)
);

create table feed_icons (
    feed_id bigint not null,
    icon_id bigint not null,
    primary key(feed_id, icon_id),
    foreign key (feed_id) references feeds(id) on delete cascade,
    foreign key (icon_id) references icons(id) on delete cascade
);
