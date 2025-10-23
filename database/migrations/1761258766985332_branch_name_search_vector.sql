alter table "branch"
add column "search_vector" tsvector generated always as (
    setweight(to_tsvector('english', "name"), 'A')
) stored;

create index if not exists "search_vector_idx" on "branch" using gin("search_vector");