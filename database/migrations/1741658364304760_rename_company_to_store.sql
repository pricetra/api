alter table company
    rename to store;

alter table branch
    rename column company_id to store_id;

alter table stock
    rename column company_id to store_id;

alter table price
    rename column company_id to store_id;
