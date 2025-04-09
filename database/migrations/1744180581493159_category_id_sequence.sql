select setval('category_id_seq', (select max("id") from "category"));
