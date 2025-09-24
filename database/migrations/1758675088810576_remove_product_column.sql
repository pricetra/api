alter table "product" drop column "search_vector";
alter table "product" drop column "color";
alter table "product" drop column "lowest_recorded_price";
alter table "product" drop column "highest_recorded_price";
alter table "product" drop column "url";
alter table "product" drop column "model";

-- Add column with updated ranking
alter table "product"
add column "search_vector" tsvector generated always as (
    setweight(
        to_tsvector('english', "name"),
        'A'
    ) || setweight(
        to_tsvector('english', "code"),
        'A'
    ) || setweight(
        to_tsvector('english', "brand"),
        'A'
    ) || setweight(
        to_tsvector('english', coalesce(cast("weight_value" as text), '')),
        'B'
    ) || setweight(
        to_tsvector('english', coalesce("description", '')),
        'C'
    )
) stored;
-- Recrate index if it does not exist
create index if not exists "search_vector_idx" on "product" using gin("search_vector");
