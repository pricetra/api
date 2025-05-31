create table "branch_list" (
    "id" bigserial unique primary key,
    "user_id" bigint references "user"("id") on delete cascade not null,
    "list_id" bigint references "list"("id") on delete cascade not null,
    "branch_id" bigint references "branch"("id") on delete cascade not null,
    "created_at" timestamp with time zone default now() not null
);
