alter table "product"
add column "quantity_value" integer default 1 not null,
add column "quantity_type" varchar(20) default 'count' not null;

create index if not exists "product_quantity_value_idx" on "product"("quantity_value");
create index if not exists "product_quantity_type_idx" on "product"("quantity_type");
