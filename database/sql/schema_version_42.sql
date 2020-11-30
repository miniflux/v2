alter table entries add column created_at timestamp with time zone not null default now();
update entries set created_at = published_at;
