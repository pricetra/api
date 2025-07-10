alter table "category"
add column "search_vector" tsvector generated always as (
        setweight(
            to_tsvector('english', coalesce("name", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("category_alias", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("expanded_pathname", '')),
            'B'
        )
    ) stored;

create index "category_search_vector_idx" on "category" using gin("search_vector");
create index "address_search_vector_idx" on "address" using gin("search_vector");
