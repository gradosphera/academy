CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE lesson_access AS ENUM (
    'unlocked', 'sequential', 'scheduled'
);
CREATE TYPE lesson_type AS ENUM (
    'video', 'audio', 'text', 'event'
);
CREATE TYPE material_category AS ENUM (
    'lesson_content', 'lesson_cover', 'materials', 'homework', -- with lesson_id
    'slides', 'tos', 'privacy_policy', -- with mini_app_id
    'bonus' -- with product_level_id
);
CREATE TYPE material_type AS ENUM (
    'circle_video', 'video', 'audio', 'picture', 'text',
    'quiz', 'open_question'
);
CREATE TYPE material_status AS ENUM (
    'ready', 'pending_compressing', 'pending_move_to_mux'
);
CREATE TYPE lesson_progress_status AS ENUM (
    'pending', 'failed', 'accepted'
);
CREATE TYPE payment_status AS ENUM (
    'pending', 'completed', 'failed', 'pending_refund', 'refunded'
);
CREATE TYPE user_role AS ENUM (
    'owner', 'moderator', 'student'
);
CREATE TYPE payment_service AS ENUM (
    'ton', 'wayforpay'
);

CREATE TABLE IF NOT EXISTS plans (
    "id" VARCHAR(100) PRIMARY KEY NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "description" TEXT NOT NULL,
    "price" DECIMAL NOT NULL,
    "full_price" DECIMAL NOT NULL,
    "currency" VARCHAR(10) NOT NULL,
    "max_total_students" BIGINT,
    "max_total_products" BIGINT,
    "max_total_events" BIGINT,
    "max_storage_size" BIGINT,
    "personalization" BOOLEAN NOT NULL,
    "tech_support" BOOLEAN NOT NULL,
    "duration" INTERVAL,
    "is_active" BOOLEAN NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO plans (
    "id", "name", "description", "price", "full_price", "currency",
    "max_total_students", "max_total_products", "max_total_events", "max_storage_size",
    "personalization", "tech_support", "duration", "is_active"
) VALUES
(
    'free_forever', 'Forever Free', '', 0, 0, '',
    20, 1, 1, 5000000000, FALSE, FALSE, NULL, TRUE
),
(
    'standard_monthly', 'Standard', '', 19, 19, 'BLG',
    500, 10, 10, 10000000000, TRUE, TRUE, '1 month', TRUE
),
(
    'standard_yearly', 'Standard', '', 144, 144, 'BLG',
    500, 10, 10, 10000000000, TRUE, TRUE, '1 year', TRUE
),
(
    'premium_monthly', 'Premium', '', 49, 49, 'BLG',
    NULL, NULL, NULL, 500000000000, TRUE, TRUE, '1 month', TRUE
),
(
    'premium_yearly', 'Premium', '', 444, 444, 'BLG',
    NULL, NULL, NULL, 500000000000, TRUE, TRUE, '1 year', TRUE
),
(
    'premium_promo', 'Promotional Premium', '', 0, 0, '',
    NULL, NULL, NULL, 500000000000, TRUE, TRUE, NULL, FALSE
);

CREATE TABLE IF NOT EXISTS mini_apps (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "plan_id" VARCHAR(100) REFERENCES plans("id") NOT NULL,
    "bot_id" BIGINT UNIQUE,
    "bot_token" VARCHAR(200) NOT NULL,
    "owner_telegram_id" BIGINT NOT NULL UNIQUE,
    "name" VARCHAR(100) NOT NULL UNIQUE,
    "logo" VARCHAR(255) NOT NULL,
    "logo_size" INT NOT NULL,
    "teacher_bio" TEXT DEFAULT '' NOT NULL,
    "teacher_links" VARCHAR(255)[] DEFAULT ARRAY[]::VARCHAR(255)[] NOT NULL,
    "teacher_avatar" VARCHAR(255) NOT NULL,
    "teacher_avatar_size" INT DEFAULT 0 NOT NULL,
    "color_theme" JSONB DEFAULT '{}' NOT NULL,
    "language" VARCHAR(100) NOT NULL,
    "payment_metadata" JSONB,
    "url" VARCHAR(255) NOT NULL,
    "support" VARCHAR(255) NOT NULL,
    "analytics" JSONB DEFAULT '{}' NOT NULL,
    "is_active" BOOLEAN NOT NULL,
    "active_payment_services" payment_service[] DEFAULT ARRAY[]::payment_service[] NOT NULL,

    "storage_size" BIGINT DEFAULT 0 NOT NULL,
    "total_products" BIGINT DEFAULT 0 NOT NULL,
    "total_students" BIGINT DEFAULT 0 NOT NULL,
    "total_events" BIGINT DEFAULT 0 NOT NULL,
    "max_storage_size" BIGINT,
    "max_total_products" BIGINT,
    "max_total_students" BIGINT,
    "max_total_events" BIGINT,

    "deleted_at" TIMESTAMP WITH TIME ZONE,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE NOT NULL,
    "index" INT NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "cover" VARCHAR(255) NOT NULL,
    "cover_size" INT NOT NULL,
    "description" TEXT NOT NULL,
    "content_type" VARCHAR(100) NOT NULL,
    "tags" VARCHAR(100)[] DEFAULT ARRAY[]::VARCHAR(100)[] NOT NULL,
    "lesson_access" lesson_access NOT NULL,
    "release_date" TIMESTAMP WITH TIME ZONE,
    "access_time" INTERVAL,
    "is_active" BOOLEAN NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS product_levels (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "product_id" UUID REFERENCES products("id") ON DELETE CASCADE NOT NULL,
    "index" INT NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "description" TEXT NOT NULL,
    "price" DECIMAL NOT NULL,
    "full_price" DECIMAL NOT NULL,
    "currency" VARCHAR(10) NOT NULL,
    "duration" INTERVAL,
    "is_active" BOOLEAN NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS lessons (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "product_id" UUID REFERENCES products("id") ON DELETE CASCADE NOT NULL,
    "index" INT NOT NULL,
    "module_name" VARCHAR(100) NOT NULL,
    "content_type" lesson_type NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "description" TEXT NOT NULL,
    "previous_lesson_id" UUID REFERENCES lessons("id") ON DELETE SET NULL,
    "release_date" TIMESTAMP WITH TIME ZONE,
    "access_time" INTERVAL,
    "is_active" BOOLEAN NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS product_level_lessons (
    "product_level_id" UUID REFERENCES product_levels("id") ON DELETE CASCADE NOT NULL,
    "lesson_id" UUID REFERENCES lessons("id") ON DELETE CASCADE NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    PRIMARY KEY("product_level_id", "lesson_id")
);

CREATE TABLE IF NOT EXISTS materials (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,

    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE,
    "lesson_id" UUID REFERENCES lessons("id") ON DELETE CASCADE,
    "product_level_id" UUID REFERENCES product_levels("id") ON DELETE CASCADE,

    "index" INT NOT NULL,
    "category" material_category NOT NULL,
    "content_type" material_type NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "description" TEXT NOT NULL,
    "url" VARCHAR(255) NOT NULL,
    "original_filename" VARCHAR(100) NOT NULL,
    "filename" VARCHAR(255) NOT NULL,
    "size" INT DEFAULT 0 NOT NULL,
    "metadata" JSONB,
    "hidden_metadata" JSONB,
    "status" material_status DEFAULT 'ready' NOT NULL,

    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    UNIQUE("mini_app_id", "category", "index"),
    UNIQUE("lesson_id", "category", "index"),
    UNIQUE("product_level_id", "category", "index"),

    CHECK (
        ("mini_app_id" IS NOT NULL AND "lesson_id" IS NULL AND "product_level_id" IS NULL) OR
        ("mini_app_id" IS NULL AND "lesson_id" IS NOT NULL AND "product_level_id" IS NULL) OR
        ("mini_app_id" IS NULL AND "lesson_id" IS NULL AND "product_level_id" IS NOT NULL)
    ),

    CHECK (
        (
            "lesson_id" IS NOT NULL AND
            "category" = 'lesson_content' AND
            "index" = 0 AND
            "content_type" IN ('circle_video', 'video', 'audio', 'picture', 'text')
        ) OR
        (
            "lesson_id" IS NOT NULL AND
            "category" = 'lesson_cover' AND
            "index" = 0 AND
            "content_type" IN ('picture')
        ) OR
        (
            "lesson_id" IS NOT NULL AND
            "category" = 'materials' AND
            "content_type" IN ('circle_video', 'video', 'audio', 'picture', 'text')
        ) OR
        (
            "lesson_id" IS NOT NULL AND
            "index" = 0 AND
            "category" = 'homework' AND
            "content_type" IN ('quiz', 'open_question')
        ) OR
        (
            "lesson_id" IS NOT NULL AND
            0 < "index" AND
            "category" = 'homework' AND
            "content_type" NOT IN ('quiz', 'open_question')
        ) OR
        (
            "mini_app_id" IS NOT NULL AND
            ("index" BETWEEN 0 AND 4) AND
            "category" = 'slides' AND
            "content_type" = 'picture'
        ) OR
        (
            "mini_app_id" IS NOT NULL AND
            "index" = 0 AND
            "category" IN ('tos', 'privacy_policy') AND
            "content_type" = 'text'
        ) OR (
            "product_level_id" IS NOT NULL AND
            "category" = 'bonus' AND
            "content_type" IN ('circle_video', 'video', 'audio', 'picture', 'text')
        )
    )
);

CREATE TABLE IF NOT EXISTS chunks (
    "material_id" UUID REFERENCES materials("id") ON DELETE CASCADE NOT NULL,
    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE NOT NULL,
    "index" INT NOT NULL,
    "size" INT DEFAULT 0 NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    PRIMARY KEY("material_id", "index")
);

CREATE TABLE IF NOT EXISTS users (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE NOT NULL,
    "role" user_role NOT NULL,
    "telegram_id" BIGINT NOT NULL,
    "telegram_username" VARCHAR(32) NOT NULL,
    "first_name" VARCHAR(100) NOT NULL,
    "last_name" VARCHAR(100) NOT NULL,
    "avatar" VARCHAR(255) NOT NULL,
    "avatar_size" INT DEFAULT 0 NOT NULL,
    "language" VARCHAR(100) NOT NULL,
    "color_theme" JSONB DEFAULT '{}' NOT NULL,
    "is_active" BOOLEAN NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    UNIQUE("mini_app_id", "telegram_id")
);

CREATE TABLE IF NOT EXISTS lesson_progress (
    "user_id" UUID REFERENCES users("id") ON DELETE CASCADE NOT NULL,
    "lesson_id" UUID REFERENCES lessons("id") ON DELETE CASCADE NOT NULL,
    "status" lesson_progress_status NOT NULL,
    "data" JSONB DEFAULT '{}',
    "score" INT NOT NULL,
    "size" INT DEFAULT 0 NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    PRIMARY KEY("lesson_id", "user_id")
);

CREATE TABLE IF NOT EXISTS payments (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE NOT NULL,
    "product_id" UUID REFERENCES products("id") ON DELETE SET NULL,
    "user_id" UUID REFERENCES users("id") ON DELETE SET NULL,
    "plan_id" VARCHAR(100) REFERENCES plans("id"),
    "product_level_id" UUID REFERENCES product_levels("id") ON DELETE SET NULL,
    "access_start" TIMESTAMP WITH TIME ZONE NOT NULL,
    "access_duration" INTERVAL,
    "amount" DECIMAL NOT NULL,
    "currency" VARCHAR(10) NOT NULL,
    "amount_blg" DECIMAL(12,2) NOT NULL,
    "status" payment_status NOT NULL,
    "url" VARCHAR(255) NOT NULL,
    "comment" TEXT DEFAULT '' NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS paid_lessons (
    "payment_id" UUID REFERENCES payments("id") ON DELETE CASCADE NOT NULL,
    "lesson_id" UUID REFERENCES lessons("id") ON DELETE CASCADE NOT NULL,

    PRIMARY KEY("payment_id", "lesson_id")
);

CREATE TABLE IF NOT EXISTS reviews (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "user_id" UUID REFERENCES users("id") ON DELETE CASCADE NOT NULL,
    "lesson_id" UUID REFERENCES lessons("id") ON DELETE CASCADE NOT NULL,
    "score" INT NOT NULL,
    "text" TEXT NOT NULL,

    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    UNIQUE("lesson_id", "user_id")
);

CREATE TABLE IF NOT EXISTS product_level_invites (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "user_id" UUID REFERENCES users("id") ON DELETE CASCADE,
    "product_level_id" UUID REFERENCES product_levels("id") ON DELETE CASCADE NOT NULL,

    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS product_access (
    "user_id" UUID REFERENCES users("id") ON DELETE CASCADE NOT NULL,
    "product_id" UUID REFERENCES products("id") ON DELETE CASCADE NOT NULL,
    
    "deleted_reason" VARCHAR(255) NOT NULL,

    "deleted_at" TIMESTAMP WITH TIME ZONE,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    PRIMARY KEY("user_id", "product_id")
);

CREATE TABLE IF NOT EXISTS permissions (
    "name" VARCHAR(60) PRIMARY KEY NOT NULL,
    "description" TEXT NOT NULL
);

INSERT INTO permissions ("name", "description") VALUES
('Student Management', 'Invite/promote/remove students in products.'),
('Products Control', 'Create/edit/delete product content.'),
('Subscription Management', 'Manage teacher''s subscription plans.'),
('Analytics', 'Full access to reports (student''s feedback, perfomance and setting G-tag, GA, and Facebook Pixel).'),
('Branding', 'Manage mini app branding.'),
('Account Settings', 'Change personal account settings.'),
('Student Interaction', 'Grade, and leave feedback for students.');

CREATE TABLE IF NOT EXISTS mod_invites (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    "mini_app_id" UUID REFERENCES mini_apps("id") ON DELETE CASCADE NOT NULL,
    "user_id" UUID REFERENCES users("id") ON DELETE CASCADE,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    UNIQUE("mini_app_id", "user_id")
);

CREATE TABLE IF NOT EXISTS mod_invite_permissions (
    "invite_id" UUID REFERENCES mod_invites("id") ON DELETE CASCADE NOT NULL,
    "permission_name" VARCHAR(60) REFERENCES permissions("name") ON DELETE CASCADE NOT NULL,

    PRIMARY KEY("invite_id", "permission_name")
);
