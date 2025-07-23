create table "grocery_list" (
    "id" bigserial unique primary key,
    "user_id" bigint not null references "user"("id") on delete cascade,
    "default" boolean not null default false,
    "name" text not null,
    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);
create unique index "grocery_list_user_id_default_key" ON "grocery_list" ("user_id", "default");

create table "grocery_list_item" (
    "id" bigserial unique primary key,
    "grocery_list_id" bigint not null references "grocery_list"("id") on delete cascade,
    "user_id" bigint not null references "user"("id") on delete cascade,
    "product_id" bigint references "product"("id") on delete set null,
    "quantity" integer not null default 1,
    "unit" text,
    "category" varchar(100),
    "weight" varchar(255),
    "completed" boolean not null default false,
    "created_at" timestamp with time zone default now() not null,
    "updated_at" timestamp with time zone default now() not null
);

create table "grocery_list_result" (
    "id" bigserial unique primary key,
    "grocery_list_id" bigint not null references "grocery_list"("id") on delete cascade,
    "user_id" bigint not null references "user"("id") on delete cascade,
    "branch_id" bigint references "branch"("id") on delete set null,
    "store_id" bigint references "store"("id") on delete set null,
    "total_price" numeric(10, 3) not null default 0,
    "currency_code" varchar(3) references "currency"("currency_code") default 'USD' not null,
    "product_ids" bigint[] not null default '{}',
    "created_at" timestamp with time zone default now() not null
);

-- Create a "Favorites" list for every user
insert into "grocery_list" ("name", "default", "user_id")
select 'Shopping List', true::boolean, "user"."id"
from "user";

-- Create function to create default grocery list
create function create_default_grocery_list() returns trigger as $$ begin -- Insert "Favorites"
insert into "grocery_list" ("name", "default", "user_id")
values (
        'Shopping List',
        true::boolean,
        new.id
    );
return new;
end;
$$ language plpgsql;
-- Create trigger
create trigger trg_create_default_grocery_list
after
insert on "user" for each row execute function create_default_grocery_list();
