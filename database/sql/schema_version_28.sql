alter table entries add column share_code text not null default '';
create unique index entries_share_code_idx on entries using btree(share_code) where share_code <> '';
