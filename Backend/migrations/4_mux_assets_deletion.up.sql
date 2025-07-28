CREATE TABLE IF NOT EXISTS mux_assets_to_delete (
    "asset_id" VARCHAR(100) PRIMARY KEY NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE FUNCTION func_mux_assets_to_delete()
RETURNS TRIGGER AS $$
BEGIN

    IF COALESCE(OLD.metadata->>'asset_id', '') = '' OR
        COALESCE(OLD.metadata->>'asset_id', '') = COALESCE(NEW.metadata->>'asset_id', '') THEN

        RETURN COALESCE(NEW, OLD);
    END IF;

    INSERT INTO mux_assets_to_delete ("asset_id")
    VALUES (OLD.metadata->>'asset_id')
    ON CONFLICT ("asset_id") DO NOTHING;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_material_mux_assets_to_delete
AFTER UPDATE OR DELETE ON materials
FOR EACH ROW
EXECUTE FUNCTION func_mux_assets_to_delete();
