create extension "uuid-ossp";

alter table "auth_state"
alter column "id" set default (uuid_generate_v4());

alter table "auth_state"
alter column "id" type uuid using "id"::uuid;
