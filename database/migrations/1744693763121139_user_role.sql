create type "user_role_type" as enum (
    'SUPER_ADMIN',
    'ADMIN',
    'CONTRIBUTOR',
    'CONSUMER'
);

alter table "user"
add column "role" "user_role_type" not null default 'CONSUMER'::"user_role_type";

create index "user_role_idx" on "user"("role");

update "user"
set "role" = 'SUPER_ADMIN'::"user_role_type"
where id = 1
    and email = 'ayaanqui.com@gmail.com';
