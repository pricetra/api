select setval('product_id_seq', (select max("id") from "product"));