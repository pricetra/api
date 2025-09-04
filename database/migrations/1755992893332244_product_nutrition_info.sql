create table "product_nutrition" (
    "product_id" bigint references "product"("id") on delete cascade primary key,
    "ingredient_text" text,
    "ingredient_list" text[],
    "nutriments" jsonb,
    "serving_size" varchar(50),
    "serving_size_value" float,
    "serving_size_unit" varchar(20),
    "openfoodfacts_updated_at" timestamp not null,
    "vegan" boolean,
    "vegetarian" boolean,
    "gluten_free" boolean,
    "lactose_free" boolean,
    "halal" boolean,
    "kosher" boolean,
    "created_at" timestamp default now() not null,
    "updated_at" timestamp default now() not null
);

create index if not exists "product_nutrition_ingredient_text_idx" on "product_nutrition"("ingredient_text");
create index if not exists "product_nutrition_ingredient_list_idx" on "product_nutrition"("ingredient_list");
create index if not exists "product_nutrition_openfoodfacts_updated_at_idx" on "product_nutrition"("openfoodfacts_updated_at");
