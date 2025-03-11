alter table store
    rename constraint company_pkey to store_pkey;

alter table branch
    rename constraint branch_company_id_fkey to branch_store_id_fkey;

alter table stock
    rename constraint stock_company_id_fkey to stock_store_id_fkey;

alter table price
    rename constraint price_company_id_fkey to price_store_id_fkey;
