CREATE TABLE "password_reset" (
    "id" bigserial unique primary key,
    "code" varchar(20) NOT NULL,
    "user_id" bigint references "user"("id") ON DELETE CASCADE NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);
