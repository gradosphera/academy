DROP TABLE IF EXISTS mod_invite_permissions;
DROP TABLE IF EXISTS mod_invites;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS product_access;
DROP TABLE IF EXISTS product_level_invites;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS paid_lessons;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS lesson_progress;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS product_level_lessons;
DROP TABLE IF EXISTS lessons;
DROP TABLE IF EXISTS product_levels;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS mini_apps;
DROP TABLE IF EXISTS plans;

DROP TYPE IF EXISTS payment_service;
DROP TYPE IF EXISTS lesson_access;
DROP TYPE IF EXISTS lesson_type;
DROP TYPE IF EXISTS material_category;
DROP TYPE IF EXISTS material_type;
DROP TYPE IF EXISTS material_status;
DROP TYPE IF EXISTS lesson_progress_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS user_role;

DROP EXTENSION IF EXISTS "uuid-ossp";
