CREATE TABLE "user" (
    "id" bigserial unique primary key,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL,
    "email" varchar(255) unique NOT NULL,
    "phone_number" varchar(15) unique,
    "name" varchar(255) NOT NULL,
    "password" text,
    "avatar" text,
    "birth_date" date,
    "active" boolean DEFAULT false NOT NULL
);

CREATE TYPE "user_auth_platform_type" AS ENUM (
  'INTERNAL', 'APPLE', 'GOOGLE'
);

CREATE TABLE "auth_state" (
    "id" bigserial unique primary key NOT NULL,
    "logged_in_at" timestamp with time zone DEFAULT now() NOT NULL,
    "user_id" bigint references "user"("id") ON DELETE CASCADE NOT NULL,
    "ip_address" varchar(39),
    "platform" "user_auth_platform_type" DEFAULT 'INTERNAL'::"user_auth_platform_type" NOT NULL
);

CREATE TABLE "email_verification" (
    "id" bigserial unique primary key,
    "code" varchar(20) NOT NULL,
    "user_id" bigint references "user"("id") ON DELETE CASCADE NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);


CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE "country_code_alpha_2" AS ENUM (
    'AF','AX','AL','DZ','AS','AD','AO','AI','AQ','AG','AR','AM','AW','AU','AT','AZ','BS','BH','BD','BB','BY','BE','BZ','BJ','BM','BT','BO','BA','BW','BV','BR','IO','BN','BG','BF','BI','KH','CM','CA','CV','KY','CF','TD','CL','CN','CX','CC','CO','KM','CG','CD','CK','CR','CI','HR','CU','CY','CZ','DK','DJ','DM','DO','EC','EG','SV','GQ','ER','EE','ET','FK','FO','FJ','FI','FR','GF','PF','TF','GA','GM','GE','DE','GH','GI','GR','GL','GD','GP','GU','GT','GG','GN','GW','GY','HT','HM','VA','HN','HK','HU','IS','IN','ID','IR','IQ','IE','IM','IL','IT','JM','JP','JE','JO','KZ','KE','KI','KR','KP','KW','KG','LA','LV','LB','LS','LR','LY','LI','LT','LU','MO','MK','MG','MW','MY','MV','ML','MT','MH','MQ','MR','MU','YT','MX','FM','MD','MC','MN','ME','MS','MA','MZ','MM','NA','NR','NP','NL','AN','NC','NZ','NI','NE','NG','NU','NF','MP','NO','OM','PK','PW','PS','PA','PG','PY','PE','PH','PN','PL','PT','PR','QA','RE','RO','RU','RW','BL','SH','KN','LC','MF','PM','VC','WS','SM','ST','SA','SN','RS','SC','SL','SG','SK','SI','SB','SO','ZA','GS','ES','LK','SD','SR','SJ','SZ','SE','CH','SY','TW','TJ','TZ','TH','TL','TG','TK','TO','TT','TN','TR','TM','TC','TV','UG','UA','AE','GB','US','UM','UY','UZ','VU','VE','VN','VG','VI','WF','EH','YE','ZM','ZW', 'SS', 'XK', 'BQ'
);

CREATE TABLE "currency" (
    "currency_code" varchar(3) unique primary key NOT NULL,
    "name" varchar(100) NOT NULL,
    "symbol" varchar(10) NOT NULL,
    "symbol_native" varchar(10) NOT NULL,
    "decimals" integer NOT NULL,
    "num_to_basic" integer
);

CREATE TABLE "country" (
    "code" "country_code_alpha_2" unique primary key NOT NULL,
    "name" varchar(56) NOT NULL,
    "currency" varchar(3) references "currency"("currency_code"),
    "calling_code" varchar(20),
    "language" varchar(3)
);

CREATE TABLE "administrative_division" (
    "country_code" "country_code_alpha_2" references "country"("code") NOT NULL,
    "administrative_division" varchar(100) NOT NULL,
    "cities" varchar(100)[] NOT NULL
);

CREATE TABLE "address" (
    "id" bigserial unique primary key NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL,
    "latitude" double precision NOT NULL,
    "longitude" double precision NOT NULL,
    "maps_link" text NOT NULL,
    "full_address" varchar(255) NOT NULL,
    "country_code" "country_code_alpha_2" DEFAULT 'US'::"country_code_alpha_2" references "country"("code") NOT NULL,
    "created_by_id" bigint,
    "updated_by_id" bigint,
    "coordinates" geography(Point,4326) GENERATED ALWAYS AS (st_setsrid(st_makepoint("longitude", "latitude"), 4326)) STORED NOT NULL,
    "administrative_division" varchar(100) DEFAULT ''::varchar NOT NULL,
    "city" varchar(100) DEFAULT ''::varchar NOT NULL
);
