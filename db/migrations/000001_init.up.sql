-- Создание схемы, если она не существует
CREATE SCHEMA IF NOT EXISTS schema_name;

BEGIN;
-- Создание таблицы внутри схемы
CREATE TABLE schema_name.orders (
                                    id serial not null,
                                    item serial not null,
                                    quantity serial not null
);

COMMIT;