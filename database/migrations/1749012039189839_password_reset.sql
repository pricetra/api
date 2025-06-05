CREATE TABLE "password_reset" (
    "id" bigserial unique primary key,
    "code" varchar(20) not null,
    "user_id" bigint references "user"("id") on delete cascade not null,
    "created_at" timestamp with time zone default now() not null,
    "tries" int not null default 0
);
