alter table "price"
add column "sale" boolean not null default false,
add column "original_price" numeric(10, 3),
add column "condition" text,
add column "unit_type" varchar(255) not null default 'item',
add column "image_id" varchar(255),
add column "expires_at" timestamptz;
