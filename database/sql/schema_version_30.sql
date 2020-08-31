alter table feeds add column next_check_at timestamp with time zone default now();
create index entries_user_feed_idx on entries (user_id, feed_id);
