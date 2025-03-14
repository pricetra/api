create type "product_source_type" as enum ('Pricetra', 'UPCitemdb', 'Other', 'Unknown');

alter table "product"
add column "source" "product_source_type" not null default 'Pricetra'::"product_source_type";

update "product" set "source" = 'UPCitemdb'::"product_source_type";
