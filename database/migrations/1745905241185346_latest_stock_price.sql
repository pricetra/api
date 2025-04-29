alter table "stock"
add column "latest_price_id" bigint references "price"("id");
