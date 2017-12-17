drop table tokens;

create table sessions (
    id text not null,
    data jsonb not null,
    created_at timestamp with time zone not null default now(),
    primary key(id)
);