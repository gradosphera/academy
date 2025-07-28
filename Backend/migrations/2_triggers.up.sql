
CREATE OR REPLACE FUNCTION check_materials_limit()
RETURNS TRIGGER AS $$
DECLARE
    max_limit INTEGER;
BEGIN
    CASE
        WHEN NEW.category = 'materials' AND NEW.filename <> '' THEN
            IF 5 <= (
                SELECT COUNT(*) FROM materials
                WHERE "category" = NEW.category AND "filename" <> '' AND (
                    (NEW.mini_app_id IS NOT NULL AND mini_app_id = NEW.mini_app_id) OR
                    (NEW.lesson_id IS NOT NULL AND lesson_id = NEW.lesson_id) OR
                    (NEW.product_level_id IS NOT NULL AND product_level_id = NEW.product_level_id)
                )
            ) THEN
                RAISE EXCEPTION 'Number of materials with file exceeds the limit';
            END IF;
        WHEN NEW.category = 'materials' AND NEW.url <> '' THEN
            IF 5 <= (
                SELECT COUNT(*) FROM materials
                WHERE "category" = NEW.category AND "url" <> '' AND (
                    (NEW.mini_app_id IS NOT NULL AND mini_app_id = NEW.mini_app_id) OR
                    (NEW.lesson_id IS NOT NULL AND lesson_id = NEW.lesson_id) OR
                    (NEW.product_level_id IS NOT NULL AND product_level_id = NEW.product_level_id)
                )
            ) THEN
                RAISE EXCEPTION 'Number of materials with link exceeds the limit';
            END IF;
        WHEN NEW.category = 'bonus' AND NEW.filename <> '' THEN
            IF 5 <= (
                SELECT COUNT(*) FROM materials
                WHERE "category" = NEW.category AND "filename" <> '' AND (
                    (NEW.mini_app_id IS NOT NULL AND mini_app_id = NEW.mini_app_id) OR
                    (NEW.lesson_id IS NOT NULL AND lesson_id = NEW.lesson_id) OR
                    (NEW.product_level_id IS NOT NULL AND product_level_id = NEW.product_level_id)
                )
            ) THEN
                RAISE EXCEPTION 'Number of bonus materials with file exceeds the limit';
            END IF;
        WHEN NEW.category = 'bonus' AND NEW.url <> '' THEN
            IF 5 <= (
                SELECT COUNT(*) FROM materials
                WHERE "category" = NEW.category AND "url" <> '' AND (
                    (NEW.mini_app_id IS NOT NULL AND mini_app_id = NEW.mini_app_id) OR
                    (NEW.lesson_id IS NOT NULL AND lesson_id = NEW.lesson_id) OR
                    (NEW.product_level_id IS NOT NULL AND product_level_id = NEW.product_level_id)
                )
            ) THEN
                RAISE EXCEPTION 'Number of bonus materials with link exceeds the limit';
            END IF;
        ELSE RETURN NEW;
    END CASE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_materials_limit
BEFORE INSERT OR UPDATE ON materials
FOR EACH ROW
EXECUTE FUNCTION check_materials_limit();

CREATE OR REPLACE FUNCTION func_account_mini_app_changes()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_OP
    WHEN 'INSERT' THEN
        NEW.storage_size = NEW.logo_size + NEW.teacher_avatar_size;
        NEW.max_storage_size = (SELECT "max_storage_size" FROM plans WHERE id = NEW.plan_id);
        NEW.max_total_products = (SELECT "max_total_products" FROM plans WHERE id = NEW.plan_id);
        NEW.max_total_students = (SELECT "max_total_students" FROM plans WHERE id = NEW.plan_id);
        NEW.max_total_events = (SELECT "max_total_events" FROM plans WHERE id = NEW.plan_id);

        IF NEW.max_storage_size < NEW.storage_size THEN
            RAISE EXCEPTION 'Storage size exceeds the limit';
        END IF;
    WHEN 'UPDATE' THEN
        IF NEW.plan_id <> OLD.plan_id THEN
            NEW.max_storage_size = (SELECT "max_storage_size" FROM plans WHERE id = NEW.plan_id);
            NEW.max_total_products = (SELECT "max_total_products" FROM plans WHERE id = NEW.plan_id);
            NEW.max_total_students = (SELECT "max_total_students" FROM plans WHERE id = NEW.plan_id);
            NEW.max_total_events = (SELECT "max_total_events" FROM plans WHERE id = NEW.plan_id);
        END IF;

        NEW.storage_size = NEW.storage_size + (NEW.logo_size - OLD.logo_size) + (NEW.teacher_avatar_size - OLD.teacher_avatar_size);

        -- check all limits here instead of every trigger that changes mini_apps table:
        
        IF OLD.storage_size < NEW.storage_size AND NEW.max_storage_size < NEW.storage_size THEN
            RAISE EXCEPTION 'Storage size exceeds the limit';
        END IF;
        
        IF OLD.total_products < NEW.total_products AND NEW.max_total_products < NEW.total_products THEN
            RAISE EXCEPTION 'Product number exceeds the limit';
        END IF;
        
        IF OLD.total_students < NEW.total_students AND NEW.max_total_students < NEW.total_students THEN
            RAISE EXCEPTION 'Total students number exceeds the limit';
        END IF;
        
        IF OLD.total_events < NEW.total_events AND NEW.max_total_events < NEW.total_events THEN
            RAISE EXCEPTION 'Number of events exceeds the limit';
        END IF;

    END CASE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mini_app_changes
BEFORE INSERT OR UPDATE ON mini_apps
FOR EACH ROW
EXECUTE FUNCTION func_account_mini_app_changes();

CREATE OR REPLACE FUNCTION func_account_product_changes()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_OP
    WHEN 'INSERT' THEN
        UPDATE mini_apps
        SET
            total_products = total_products + 1,
            storage_size = storage_size + NEW.cover_size,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.mini_app_id;
    WHEN 'UPDATE' THEN
        UPDATE mini_apps
        SET
            storage_size = storage_size + (NEW.cover_size - OLD.cover_size),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.mini_app_id;
    WHEN 'DELETE' THEN
        UPDATE mini_apps
        SET
            total_products = total_products - 1,
            storage_size = storage_size - OLD.cover_size,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = OLD.mini_app_id;
    END CASE;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_product_changes
AFTER INSERT OR UPDATE OR DELETE ON products
FOR EACH ROW
EXECUTE FUNCTION func_account_product_changes();

CREATE OR REPLACE FUNCTION func_account_lesson_changes()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE mini_apps
    SET
        total_events = total_events
            + CASE WHEN (NEW.release_date IS NOT NULL) THEN 1 ELSE 0 END
            - CASE WHEN (OLD.release_date IS NOT NULL) THEN 1 ELSE 0 END,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = (
        SELECT mini_app_id FROM products
        WHERE id = COALESCE(NEW.product_id, OLD.product_id)
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_lesson_changes
AFTER INSERT OR UPDATE OR DELETE ON lessons
FOR EACH ROW
EXECUTE FUNCTION func_account_lesson_changes();

CREATE OR REPLACE FUNCTION func_account_material_changes()
RETURNS TRIGGER AS $$
DECLARE
    affected_mini_app_id UUID;
BEGIN
    CASE
        WHEN COALESCE(NEW.mini_app_id, OLD.mini_app_id) IS NOT NULL THEN 
            affected_mini_app_id := COALESCE(NEW.mini_app_id, OLD.mini_app_id);
        WHEN COALESCE(NEW.product_level_id, OLD.product_level_id) IS NOT NULL THEN
            affected_mini_app_id := (
                SELECT DISTINCT mini_app_id FROM products WHERE id = (
                    SELECT DISTINCT product_id FROM product_levels
                    WHERE id = COALESCE(NEW.product_level_id, OLD.product_level_id)
                )
            );
        WHEN COALESCE(NEW.lesson_id, OLD.lesson_id) IS NOT NULL THEN
            affected_mini_app_id := (
                SELECT DISTINCT mini_app_id FROM products WHERE id = (
                    SELECT DISTINCT product_id FROM lessons
                    WHERE id = COALESCE(NEW.lesson_id, OLD.lesson_id)
                )
            );
    END CASE;

    UPDATE mini_apps
    SET
        storage_size = storage_size + COALESCE(NEW.size, 0) - COALESCE(OLD.size, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = affected_mini_app_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_material_changes
AFTER INSERT OR UPDATE OR DELETE ON materials
FOR EACH ROW
EXECUTE FUNCTION func_account_material_changes();

CREATE OR REPLACE FUNCTION func_account_chunk_changes()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE mini_apps
    SET
        storage_size = storage_size + COALESCE(NEW.size, 0) - COALESCE(OLD.size, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.mini_app_id, OLD.mini_app_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_chunk_changes
AFTER INSERT OR UPDATE OR DELETE ON chunks
FOR EACH ROW
EXECUTE FUNCTION func_account_chunk_changes();

CREATE OR REPLACE FUNCTION func_account_lesson_progress_changes()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE mini_apps
    SET
        storage_size = storage_size + COALESCE(NEW.size, 0) - COALESCE(OLD.size, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = (
        SELECT DISTINCT mini_app_id FROM products WHERE id = (
            SELECT DISTINCT product_id FROM lessons WHERE id = COALESCE(NEW.lesson_id, OLD.lesson_id)
        )
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_lesson_progress_changes
AFTER INSERT OR UPDATE OR DELETE ON lesson_progress
FOR EACH ROW
EXECUTE FUNCTION func_account_lesson_progress_changes();

CREATE OR REPLACE FUNCTION func_account_user_changes()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE mini_apps
    SET
        total_students = total_students
            + (CASE WHEN (NEW.role = 'student') THEN 1 ELSE 0 END)
            - (CASE WHEN (OLD.role = 'student') THEN 1 ELSE 0 END),
        storage_size = storage_size + COALESCE(NEW.avatar_size, 0) - COALESCE(OLD.avatar_size, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.mini_app_id, OLD.mini_app_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_user_changes
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION func_account_user_changes();

CREATE OR REPLACE FUNCTION func_add_paid_lessons()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.product_level_id IS NOT NULL THEN
        INSERT INTO paid_lessons ("payment_id", "lesson_id")
        SELECT NEW.id, pll.lesson_id FROM product_level_lessons pll
        WHERE pll.product_level_id = NEW.product_level_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_add_paid_lessons
AFTER INSERT ON payments
FOR EACH ROW
EXECUTE FUNCTION func_add_paid_lessons();
