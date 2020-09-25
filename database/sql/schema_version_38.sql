alter table integrations add column telegram_enabled bool default 'f';
alter table integrations add column telegram_token text default '';
alter table integrations add column telegram_chat_id text default '';