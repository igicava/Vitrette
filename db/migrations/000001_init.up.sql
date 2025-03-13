create schema if not exists lyceum_schema;

create table if not exists lyceum_schema.orders (
    id text not null,
    item text not null,
    quantity bigserial not null
)
