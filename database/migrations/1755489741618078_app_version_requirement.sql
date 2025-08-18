create table "app_version_requirement" (
    "platform" "auth_device_type" unique primary key,
    "min_version" text not null
);

insert into "app_version_requirement" ("platform", "min_version") values
    ('ios', '1.0.20'),
    ('android', '1.0.20');
