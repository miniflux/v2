create extension if not exists hstore;
alter table users add column extra hstore;
create index users_extra_idx on users using gin(extra);
