alter table "auth_state"
alter column "id" type varchar(255) using ("id"::varchar(255));
