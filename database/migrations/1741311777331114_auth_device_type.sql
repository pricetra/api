CREATE TYPE "auth_device_type" AS ENUM(
    'ios',
    'android',
    'web',
    'other',
    'unknown'
);

alter table "auth_state"
add column "device_type" "auth_device_type" default 'unknown'::"auth_device_type";
