create table "product_view" (
    "id" bigserial unique primary key,
    "product_id" bigint references "product"("id") on delete cascade not null,
    "stock_id" bigint references "stock"("id") on delete set null,
    "user_id" bigint references "user"("id") on delete set null,
    "origin" text,
    "platform" "auth_device_type" default 'other'::"auth_device_type" not null,
    "created_at" timestamp with time zone default now() not null
);

alter table "product"
add column "views" bigint default 0 not null;
