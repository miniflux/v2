alter table entries add column changed_at timestamp with time zone;
update entries set changed_at = published_at;
alter table entries alter column changed_at set not null;
