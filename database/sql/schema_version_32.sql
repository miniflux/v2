-- The following command is easier, but not working for postgres 10.
-- ALTER TYPE entry_status ADD VALUE 'marked';

-- Use the alternative method for older postgres.
-- rename the old enum
alter type entry_status rename to entry_status__;
-- create the new enum
create type entry_status as enum('unread', 'read', 'removed', 'marked');

-- alter all you enum columns
ALTER TABLE entries 
  ALTER COLUMN status DROP DEFAULT, 
  ALTER COLUMN status TYPE entry_status USING (status::text::entry_status), 
  ALTER COLUMN status SET DEFAULT 'unread';

-- drop the old enum
drop type entry_status__;
