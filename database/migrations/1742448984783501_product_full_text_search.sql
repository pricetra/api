-- Add search_vector column with full text search functionality
alter table "product"
add column "search_vector" tsvector generated always as (
        setweight(
            to_tsvector('english', coalesce("name", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("description", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("brand", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("code", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("category", '')),
            'A'
        ) || setweight(
            to_tsvector('english', coalesce("model", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("weight", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("color", '')),
            'B'
        ) || setweight(
            to_tsvector('english', coalesce("url", '')),
            'C'
        )
    ) stored;

-- Create GIN index on the search_vector column
create index "search_vector_idx" on "product" using gin("search_vector");

-- Create function to update search_vector on insert or update
create or replace function product_search_vector_update() returns trigger as $$ begin new.search_vector := setweight(
        to_tsvector('english', coalesce(new."name", '')),
        'A'
    ) || setweight(
        to_tsvector('english', coalesce(new."description", '')),
        'A'
    ) || setweight(
        to_tsvector('english', coalesce(new."brand", '')),
        'A'
    ) || setweight(
        to_tsvector('english', coalesce(new."code", '')),
        'B'
    ) || setweight(
        to_tsvector('english', coalesce(new."category", '')),
        'A'
    ) || setweight(
        to_tsvector('english', coalesce(new."model", '')),
        'B'
    ) || setweight(
        to_tsvector('english', coalesce(new."weight", '')),
        'B'
    ) || setweight(
        to_tsvector('english', coalesce(new."color", '')),
        'B'
    ) || setweight(
        to_tsvector('english', coalesce(new."url", '')),
        'C'
    );
return new;
end $$ language plpgsql;

-- Create trigger to automatically update search_vector on insert and update
create trigger update_product_search_vector before
insert
    or
update on "product" for each row execute function product_search_vector_update();
