alter table "product"
drop column if exists "weight";

create index if not exists "product_weight_value_idx" on "product"("weight_value");
create index if not exists "product_weight_type_idx" on "product"("weight_type");
