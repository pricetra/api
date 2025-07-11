create table "search_history" (
    "id" bigserial unique primary key,
    "search_term" text not null,
    "user_id" bigint references "user"("id") on delete set null,
    "created_at" timestamp with time zone default now() not null
);
