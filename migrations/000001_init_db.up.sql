create table public.users
(
    id       bigserial
        constraint users_pk
            primary key,
    login    varchar(255) not null,
    password varchar(255) not null
);

create unique index users_login_udx
    on public.users (login, login);

create table public.orders
(
    id         bigserial
        constraint orders_pk
            primary key,
    created_at timestamp with time zone default now() not null,
    number     varchar(100)                           not null
        constraint orders_number_udx
            unique,
    status     integer,
    user_id    bigint
        constraint orders_user_id_idx
            references public.users
);

create table public.balance
(
    id           bigserial
        constraint balance_pk
            primary key,
    created_at   timestamp with time zone not null,
    order_number varchar(50)              not null,
    user_id      bigint                   not null
        constraint balance_user_id_fk
            references public.users,
    sum          numeric(15, 2)           not null,
    type         smallint                 not null
);

create index balance_user_id_idx
    on public.balance (user_id);
