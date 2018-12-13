create table medias (
	id bigserial not null,
	url_hash text not null unique,
	mime_type text not null,
	content bytea not null,
	primary key (id)
);
CREATE TABLE entry_medias (
    entry_id int8 NOT NULL,
    media_id int8 NOT NULL,
    PRIMARY KEY (entry_id, media_id),
    foreign key (entry_id) references entries(id) on delete cascade,
    foreign key (media_id) references medias(id) on delete cascade
);