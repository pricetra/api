create table "company" (
    "id" bigserial unique primary key,
    "name" varchar(255) not null,
    "logo" text not null,
    "website" text not null,

    "created_by_id" bigint references "user"("id") on delete set null,
    "updated_by_id" bigint references "user"("id") on delete set null,

    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);

create table "branch" (
    "id" bigserial unique primary key,
    "name" varchar(255) not null,
    "address_id" bigint references "address"("id") on delete cascade not null,
    "company_id" bigint references "company"("id") on delete cascade not null,

    "created_by_id" bigint references "user"("id") on delete set null,
    "updated_by_id" bigint references "user"("id") on delete set null,

    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);

create table "product" (
    "id" bigserial unique primary key,
    "name" varchar(255) not null,
    "image" text not null,
    "description" text default '' not null,
    "url" text,
    "brand" varchar(255) not null,
    "code" varchar(255) unique not null,

    "created_by_id" bigint references "user"("id") on delete set null,
    "updated_by_id" bigint references "user"("id") on delete set null,

    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);

create table "stock" (
    "id" bigserial unique primary key,
    "product_id" bigint references "product"("id") on delete cascade not null,
    "company_id" bigint references "company"("id") on delete cascade not null,
    "branch_id" bigint references "branch"("id") on delete cascade not null,

    "created_by_id" bigint references "user"("id") on delete set null,
    "updated_by_id" bigint references "user"("id") on delete set null,

    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);

create table "price" (
    "id" bigserial unique primary key,
    "amount" numeric(9, 2) not null,
    "currency_code" varchar(3) references "currency"("currency_code") default 'USD' not null,
    "product_id" bigint references "product"("id") on delete cascade not null,
    "company_id" bigint references "company"("id") on delete cascade not null,
    "branch_id" bigint references "branch"("id") on delete cascade not null,
    "stock_id" bigint references "stock"("id") on delete cascade not null,

    "created_by_id" bigint references "user"("id") on delete set null,
    "updated_by_id" bigint references "user"("id") on delete set null,

    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);
