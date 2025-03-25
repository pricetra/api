-- Drop trigger and function
drop trigger if exists update_product_search_vector on "product";
drop function if exists product_search_vector_update();

-- Remove column
alter table "product" drop column if exists "search_vector";

-- Add column with updated ranking
alter table "product"
add column "search_vector" tsvector generated always as (
        setweight(
            to_tsvector('english', coalesce("name", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("brand", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("category", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("description", '')),
            'C'
        ) || setweight(
            to_tsvector('english', coalesce("url", '')),
            'C'
        ) || setweight(
            to_tsvector('english', coalesce("code", '')),
            'D'
        ) || setweight(
            to_tsvector('english', coalesce("model", '')),
            'D'
        ) || setweight(
            to_tsvector('english', coalesce("weight", '')),
            'D'
        ) || setweight(
            to_tsvector('english', coalesce("color", '')),
            'D'
        )
    ) stored;

-- Recrate index if it does not exist
create index if not exists "search_vector_idx" on "product" using gin("search_vector");
