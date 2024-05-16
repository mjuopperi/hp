create table health_data (
    id          bigserial primary key,
    timestamp   timestamp with time zone not null,
    measurement varchar(255)             not null,
    value       numeric                  not null,
    unit        varchar(255)             not null
);
