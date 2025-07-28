DROP TRIGGER IF EXISTS trg_add_paid_lessons ON payments;
DROP FUNCTION IF EXISTS func_add_paid_lessons();

DROP TRIGGER IF EXISTS trg_user_changes ON users;
DROP FUNCTION IF EXISTS func_account_user_changes();

DROP TRIGGER IF EXISTS trg_lesson_progress_changes ON lesson_progress;
DROP FUNCTION IF EXISTS func_account_lesson_progress_changes();

DROP TRIGGER IF EXISTS trg_chunk_changes ON chunks;
DROP FUNCTION IF EXISTS func_account_chunk_changes();

DROP TRIGGER IF EXISTS trg_material_changes ON materials;
DROP FUNCTION IF EXISTS func_account_material_changes();

DROP TRIGGER IF EXISTS trg_lesson_changes ON lessons;
DROP FUNCTION IF EXISTS func_account_lesson_changes();

DROP TRIGGER IF EXISTS trg_product_changes ON products;
DROP FUNCTION IF EXISTS func_account_product_changes();

DROP TRIGGER IF EXISTS trg_mini_app_changes ON mini_apps;
DROP FUNCTION IF EXISTS func_account_mini_app_changes();

DROP TRIGGER IF EXISTS enforce_materials_limit ON materials;
DROP FUNCTION IF EXISTS check_materials_limit();
