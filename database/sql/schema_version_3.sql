create table tokens (
    id text not null,
    value text not null,
    created_at timestamp with time zone not null default now(),
    primary key(id, value)
);