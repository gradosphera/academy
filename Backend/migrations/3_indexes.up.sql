CREATE INDEX IF NOT EXISTS idx_mini_apps_owner_telegram_id ON mini_apps USING HASH ("owner_telegram_id");
CREATE INDEX IF NOT EXISTS idx_mini_apps_name ON mini_apps USING HASH ("name");
CREATE INDEX IF NOT EXISTS idx_mini_apps_deleted_at ON mini_apps ("deleted_at");

CREATE INDEX IF NOT EXISTS idx_products_mini_app_id ON products USING HASH ("mini_app_id");

CREATE INDEX IF NOT EXISTS idx_product_levels_product_id ON product_levels USING HASH ("product_id");
CREATE INDEX IF NOT EXISTS idx_product_levels_duration ON product_levels ("duration");

CREATE INDEX IF NOT EXISTS idx_lessons_product_id ON lessons USING HASH ("product_id");

CREATE INDEX IF NOT EXISTS idx_product_level_lessons_lesson_id ON product_level_lessons USING HASH ("lesson_id");

CREATE INDEX IF NOT EXISTS idx_materials_mini_app_id ON materials USING HASH ("mini_app_id");
CREATE INDEX IF NOT EXISTS idx_materials_lesson_id ON materials USING HASH ("lesson_id");
CREATE INDEX IF NOT EXISTS idx_materials_product_level_id ON materials USING HASH ("product_level_id");
CREATE INDEX IF NOT EXISTS idx_materials_category ON materials USING HASH ("category");
CREATE INDEX IF NOT EXISTS idx_materials_content_type ON materials USING HASH ("content_type");
CREATE INDEX IF NOT EXISTS idx_materials_filename ON materials USING HASH ("filename");

CREATE INDEX IF NOT EXISTS idx_chunks_mini_app_id ON chunks USING HASH ("mini_app_id");
CREATE INDEX IF NOT EXISTS idx_chunks_created_at ON chunks ("created_at");

CREATE INDEX IF NOT EXISTS idx_users_mini_app_id ON users USING HASH ("mini_app_id");
CREATE INDEX IF NOT EXISTS idx_users_role ON users USING HASH ("role");
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users USING HASH ("telegram_id");

CREATE INDEX IF NOT EXISTS idx_lesson_progress_user_id ON lesson_progress USING HASH ("user_id");
CREATE INDEX IF NOT EXISTS idx_lesson_progress_status ON lesson_progress USING HASH ("status");

CREATE INDEX IF NOT EXISTS idx_payments_mini_app_id ON payments USING HASH ("mini_app_id");
CREATE INDEX IF NOT EXISTS idx_payments_product_id ON payments USING HASH ("product_id");
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments USING HASH ("user_id");
CREATE INDEX IF NOT EXISTS idx_payments_plan_id ON payments USING HASH ("plan_id");
CREATE INDEX IF NOT EXISTS idx_payments_product_level_id ON payments USING HASH ("product_level_id");
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments USING HASH ("status");

CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews USING HASH ("user_id");
CREATE INDEX IF NOT EXISTS idx_reviews_lesson_id ON reviews USING HASH ("lesson_id");

CREATE INDEX IF NOT EXISTS idx_product_level_invites_user_id ON product_level_invites USING HASH ("user_id");
CREATE INDEX IF NOT EXISTS idx_product_level_invites_product_level_id ON product_level_invites USING HASH ("product_level_id");

CREATE INDEX IF NOT EXISTS idx_product_access_product_id ON product_access USING HASH ("product_id");
CREATE INDEX IF NOT EXISTS idx_product_access_not_deleted ON product_access ("deleted_at") WHERE "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS idx_mod_invites_mini_app_id ON mod_invites USING HASH ("mini_app_id");
CREATE INDEX IF NOT EXISTS idx_mod_invites_user_id ON mod_invites USING HASH ("user_id");

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_users_telegram_username_trgm ON users USING gin ("telegram_username" gin_trgm_ops);
