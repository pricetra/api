alter table "auth_state"
add column "expo_push_token" text unique;

create table "push_notification" (
    "id" bigserial unique primary key,
    "request" jsonb,
    "response" jsonb,
    "created_at" timestamp with time zone default now() not null
);
