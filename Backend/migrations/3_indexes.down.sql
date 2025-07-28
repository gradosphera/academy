DROP INDEX IF EXISTS idx_mini_apps_owner_telegram_id;
DROP INDEX IF EXISTS idx_mini_apps_name;
DROP INDEX IF EXISTS idx_mini_apps_deleted_at;

DROP INDEX IF EXISTS idx_products_mini_app_id;
DROP INDEX IF EXISTS idx_product_levels_product_id;
DROP INDEX IF EXISTS idx_product_levels_duration;
DROP INDEX IF EXISTS idx_lessons_product_id;
DROP INDEX IF EXISTS idx_product_level_lessons_lesson_id;

DROP INDEX IF EXISTS idx_materials_mini_app_id;
DROP INDEX IF EXISTS idx_materials_lesson_id;
DROP INDEX IF EXISTS idx_materials_product_level_id;
DROP INDEX IF EXISTS idx_materials_category;
DROP INDEX IF EXISTS idx_materials_content_type;
DROP INDEX IF EXISTS idx_materials_filename;

DROP INDEX IF EXISTS idx_chunks_mini_app_id;
DROP INDEX IF EXISTS idx_chunks_created_at;
DROP INDEX IF EXISTS idx_users_mini_app_id;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP INDEX IF EXISTS idx_lesson_progress_user_id;
DROP INDEX IF EXISTS idx_lesson_progress_status;

DROP INDEX IF EXISTS idx_payments_mini_app_id;
DROP INDEX IF EXISTS idx_payments_product_id;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_plan_id;
DROP INDEX IF EXISTS idx_payments_product_level_id;
DROP INDEX IF EXISTS idx_payments_status;

DROP INDEX IF EXISTS idx_reviews_user_id;
DROP INDEX IF EXISTS idx_reviews_lesson_id;
DROP INDEX IF EXISTS idx_product_level_invites_user_id;
DROP INDEX IF EXISTS idx_product_level_invites_product_level_id;
DROP INDEX IF EXISTS idx_product_access_product_id;
DROP INDEX IF EXISTS idx_product_access_not_deleted;
DROP INDEX IF EXISTS idx_mod_invites_mini_app_id;
DROP INDEX IF EXISTS idx_mod_invites_user_id;

DROP INDEX IF EXISTS idx_users_telegram_username_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
