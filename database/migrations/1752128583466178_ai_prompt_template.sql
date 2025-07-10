create type "ai_prompt_type" as enum (
    'PRODUCT_DETAILS',
    'RECEIPT',
    'NUTRITION'
);

create table "ai_prompt_template" (
    "type" "ai_prompt_type" primary key unique not null,
    "prompt" text not null,
    "variable" varchar(255) not null,
    "max_tokens" integer not null
);

insert into "ai_prompt_template" ("type", "prompt", "variable", "max_tokens") VALUES(
    'PRODUCT_DETAILS'::"ai_prompt_type",
    'Extract brand, product name/description, weight (x lb, fl oz, etc.), and category (Google product taxonomy format) from this string: "{{ocr_string}}". Respond with a single JSON object only, using this schema: `{"brand":string,"productName":string,"weight"?:string,"category":string}`.'::text,
    '{{ocr_string}}',
    300
);

create table "ai_prompt_response" (
    "id" bigserial unique primary key,
    "type" "ai_prompt_type" not null,
    "request" jsonb,
    "response" jsonb,
    "created_at" timestamp with time zone default now() not null
);
