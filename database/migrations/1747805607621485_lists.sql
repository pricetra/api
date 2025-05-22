create type "list_type" as enum ('FAVORITES', 'WATCH_LIST', 'PERSONAL');

create table "list" (
    "id" bigserial unique primary key,
    "name" varchar(255) not null,
    "type" "list_type" not null,
    "user_id" bigint references "user"("id") on delete cascade not null,
    "created_at" timestamp with time zone default now() not null
);

-- Create a "Favorites" list for every user
insert into "list" ("name", "type", "user_id")
select 'Favorites', 'FAVORITES'::"list_type", "user"."id"
from "user";

-- Create a "Watch List" list for every user
insert into "list" ("name", "type", "user_id")
select 'Watch List', 'WATCH_LIST'::"list_type", "user"."id"
from "user";


-- Create function to create default lists
create function create_default_lists() returns trigger as $$ begin 
-- Insert "Favorites"
insert into list (name, type, user_id)
values (
        'Favorites',
        'FAVORITES'::"list_type",
        new.id
    );
-- Insert "Watch List"
insert into list (name, type, user_id)
values (
        'Watch List',
        'WATCH_LIST'::"list_type",
        new.id
    );
return new;
end;
$$ language plpgsql;

-- Create trigger
create trigger trg_create_default_lists
after
insert on "user" for each row execute function create_default_lists();


--
-- Product and Branch Lists
--

create table "product_list" (
    "id" bigserial unique primary key,
    "user_id" bigint references "user"("id") on delete cascade not null,
    "list_id" bigint references "list"("id") on delete cascade not null,
    "product_id" bigint references "product"("id") on delete cascade not null,
    "stock_id" bigint references "stock"("id") on delete set null,
    "created_at" timestamp with time zone default now() not null
);
