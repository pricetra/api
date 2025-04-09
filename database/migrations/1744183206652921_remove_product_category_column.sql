alter table "product"
drop column "category";

-- Recreate search vector
alter table "product"
add column if not exists "search_vector" tsvector generated always as (
        setweight(
            to_tsvector('english', coalesce("name", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("brand", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("description", '')),
            'C'
        )
    ) stored;
-- Recrate index if it does not exist
create index if not exists "search_vector_idx" on "product" using gin("search_vector");
