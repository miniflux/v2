create table webpush_subscriptions (
    id bigserial not null,
    user_id int not null,
    subscription text not null,
    primary key (id),
    foreign key (user_id) references users(id) on delete cascade
);