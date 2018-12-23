create table medias (
	id bigserial not null,
	url text not null,
	url_hash text not null unique,
	mime_type text default '',
	content bytea default E''::bytea,
	size int8 default 0,
	success bool default 'f',
	created_at timestamp with time zone default current_timestamp,
	primary key (id)
);
CREATE TABLE entry_medias (
    entry_id int8 NOT NULL,
    media_id int8 NOT NULL,
	use_cache bool default 'f',
    PRIMARY KEY (entry_id, media_id),
    foreign key (entry_id) references entries(id) on delete cascade,
    foreign key (media_id) references medias(id) on delete cascade
);
alter table feeds add column cache_media bool default 'f';