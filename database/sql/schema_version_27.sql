create table api_keys (
    id serial not null,
    user_id int not null references users(id) on delete cascade,
    token text not null unique,
    description text not null,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone default now(),
    primary key(id),
    unique (user_id, description)
);
