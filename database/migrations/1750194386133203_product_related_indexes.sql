-- Product indexes
create index if not exists "product_brand_idx" on "product"("brand");
create index if not exists "product_category_id_idx" on "product"("category_id");
create index if not exists "product_views_idx" on "product"("views");
create index if not exists "product_created_at_idx" on "product"("created_at");
create index if not exists "product_updated_at_idx" on "product"("updated_at");

-- Stock indexes
create index if not exists "stock_product_id_idx" on "stock"("product_id");
create index if not exists "stock_store_id_idx" on "stock"("store_id");
create index if not exists "stock_branch_id_idx" on "stock"("branch_id");
create index if not exists "stock_latest_price_id_idx" on "stock"("latest_price_id");

-- Branch indexes
create index if not exists "branch_store_id_idx" on "branch"("store_id");
create index if not exists "branch_address_id_idx" on "branch"("address_id");

-- Price indexes
create index if not exists "price_amount_idx" on "price"("amount");
create index if not exists "price_product_id_idx" on "price"("product_id");
create index if not exists "price_store_id_idx" on "price"("store_id");
create index if not exists "price_branch_id_idx" on "price"("branch_id");
create index if not exists "price_stock_id_idx" on "price"("stock_id");
create index if not exists "price_sale_idx" on "price"("sale");
create index if not exists "price_created_at_idx" on "price"("created_at");
