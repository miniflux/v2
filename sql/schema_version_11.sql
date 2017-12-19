alter table integrations add column wallabag_enabled bool default 'f';
alter table integrations add column wallabag_url text default '';
alter table integrations add column wallabag_client_id text default '';
alter table integrations add column wallabag_client_secret text default '';
alter table integrations add column wallabag_username text default '';
alter table integrations add column wallabag_password text default '';