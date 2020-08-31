create index enclosures_user_entry_url_idx on enclosures(user_id, entry_id, md5(url));
