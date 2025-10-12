create table urls(
    id serial primary key,
    alias text not null unique,
    original text not null
);

create table redirects(
    id serial primary key,
    alias text not null,
    dt timestamp not null,
    user_agent text not null
);