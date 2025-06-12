alter table "user"
add column "address_id" bigint references "address"("id") on delete set null;
