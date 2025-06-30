alter table "address"
add column "street" varchar(100);

update "address"
set "street" = (regexp_split_to_array(full_address, ','))[1]
where array_length(regexp_split_to_array(full_address, ','), 1) = 4;

create extension if not exists unaccent;

alter table "address"
add column "search_vector" tsvector generated always as (
        setweight(
            to_tsvector('english', coalesce("full_address", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("city", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("street", '')),
            'B'
        ) || setweight(
            to_tsvector('english', lpad(zip_code::text, 5, '0')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("administrative_division", '')),
            'C'
        )
    ) stored;

create index if not exists "search_vector_idx" on "address" using gin("search_vector");
