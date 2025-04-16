create type "product_billing_type" as enum ('CREATE', 'UPDATE', 'SCAN');

create table "product_billing_rate" (
    "type" "product_billing_type" unique primary key,
    "rate" numeric(3, 2) not null,
    "currency_code" varchar(3) references "currency"("currency_code") default 'USD' not null
);

insert into "product_billing_rate" ("type", "rate", "currency_code") values
    ('CREATE'::"product_billing_type", 0.40, 'USD'),
    ('UPDATE'::"product_billing_type", 0.20, 'USD'),
    ('SCAN'::"product_billing_type", 0.10, 'USD');

create table "product_billing" (
    "id" bigserial unique primary key,
    "product_id" bigint references "product"("id") on delete cascade not null,
    "user_id" bigint references "user"("id") on delete cascade not null,
    "created_at" timestamp with time zone default now() not null,
    "rate" numeric(3, 2) not null,
    "billing_rate_type" "product_billing_type" references "product_billing_rate"("type") on delete cascade not null,
    "new_data" jsonb,
    "old_data" jsonb,
    "paid_at" timestamp with time zone
);
