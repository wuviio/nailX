-- nailX 初期スキーマ
-- ================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector; -- pgvector (512次元ベクトル)

-- ----------------------------------------------------------------
-- updated_at 自動更新トリガー関数
-- ----------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------
-- users
-- ----------------------------------------------------------------
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid    VARCHAR(128) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    gender          VARCHAR(30),
    role            VARCHAR(20) NOT NULL DEFAULT 'consumer',
    avatar_url      TEXT,
    bio             TEXT,
    lifestyle_tags  TEXT[] DEFAULT '{}',
    point_balance   INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_role ON users(role);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ----------------------------------------------------------------
-- nail_profiles
-- ----------------------------------------------------------------
CREATE TABLE nail_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    avg_nail_length_mm  NUMERIC(4,1),
    nail_shape          VARCHAR(30),
    gel_lift_tendency   VARCHAR(20),
    allergy_notes       TEXT,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE TRIGGER trg_nail_profiles_updated_at
    BEFORE UPDATE ON nail_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ----------------------------------------------------------------
-- ar_sessions
-- ----------------------------------------------------------------
CREATE TABLE ar_sessions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    design_ip_id                UUID,
    detected_nail_length_mm     NUMERIC(4,1),
    has_existing_gel            BOOLEAN,
    detected_nail_shape         VARCHAR(30),
    estimated_treatment_min     INTEGER,
    estimated_gel_amount_ml     NUMERIC(5,3),
    hand_snapshot_url           TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ar_sessions_user_id ON ar_sessions(user_id);
CREATE INDEX idx_ar_sessions_created_at ON ar_sessions(created_at DESC);

-- ----------------------------------------------------------------
-- materials
-- ----------------------------------------------------------------
CREATE TABLE materials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    category        VARCHAR(50) NOT NULL,
    texture_type    VARCHAR(50),
    thumbnail_url   TEXT NOT NULL,
    model_3d_url    TEXT,
    texture_params  JSONB NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_materials_category ON materials(category);
CREATE INDEX idx_materials_is_active ON materials(is_active);

-- ----------------------------------------------------------------
-- design_ips
-- ----------------------------------------------------------------
CREATE TABLE design_ips (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id          UUID NOT NULL REFERENCES users(id),
    parent_ip_id        UUID REFERENCES design_ips(id) ON DELETE SET NULL,
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    preview_image_url   TEXT NOT NULL,
    design_data         JSONB NOT NULL,
    similarity_vector   vector(512),
    similarity_hash     VARCHAR(64),
    fork_depth          INTEGER NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    is_public           BOOLEAN NOT NULL DEFAULT true,
    gender_tag          VARCHAR(20) NOT NULL DEFAULT 'neutral',
    style_tags          TEXT[] DEFAULT '{}',
    usage_count         INTEGER NOT NULL DEFAULT 0,
    total_royalty_yen   INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_design_ips_similarity_vector ON design_ips USING hnsw (similarity_vector vector_cosine_ops);
CREATE INDEX idx_design_ips_creator_id ON design_ips(creator_id);
CREATE INDEX idx_design_ips_status ON design_ips(status);
CREATE INDEX idx_design_ips_gender_tag ON design_ips(gender_tag);
CREATE INDEX idx_design_ips_parent_ip_id ON design_ips(parent_ip_id);
CREATE INDEX idx_design_ips_usage_count ON design_ips(usage_count DESC);

CREATE TRIGGER trg_design_ips_updated_at
    BEFORE UPDATE ON design_ips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ----------------------------------------------------------------
-- design_royalty_nodes
-- ----------------------------------------------------------------
CREATE TABLE design_royalty_nodes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    design_ip_id    UUID NOT NULL REFERENCES design_ips(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    share_percent   NUMERIC(5,2) NOT NULL,
    depth_level     INTEGER NOT NULL,
    UNIQUE(design_ip_id, user_id)
);

CREATE INDEX idx_royalty_nodes_design_ip_id ON design_royalty_nodes(design_ip_id);

-- ----------------------------------------------------------------
-- salons
-- ----------------------------------------------------------------
CREATE TABLE salons (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id                UUID NOT NULL REFERENCES users(id),
    name                    VARCHAR(200) NOT NULL,
    description             TEXT,
    address                 TEXT NOT NULL,
    prefecture              VARCHAR(20) NOT NULL,
    city                    VARCHAR(50),
    lat                     NUMERIC(10,7),
    lng                     NUMERIC(10,7),
    phone                   VARCHAR(20),
    business_hours          JSONB,
    avg_reproduction_score  NUMERIC(3,2) NOT NULL DEFAULT 0.00,
    skill_badge_tags        TEXT[] DEFAULT '{}',
    portfolio_image_urls    TEXT[] DEFAULT '{}',
    verification_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    is_active               BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_salons_prefecture ON salons(prefecture);
CREATE INDEX idx_salons_verification_status ON salons(verification_status);
CREATE INDEX idx_salons_avg_reproduction_score ON salons(avg_reproduction_score DESC);
CREATE INDEX idx_salons_lat_lng ON salons(lat, lng);

CREATE TRIGGER trg_salons_updated_at
    BEFORE UPDATE ON salons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ----------------------------------------------------------------
-- booking_requests
-- ----------------------------------------------------------------
CREATE TABLE booking_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    design_ip_id        UUID NOT NULL REFERENCES design_ips(id),
    ar_session_id       UUID REFERENCES ar_sessions(id),
    nail_data_snapshot  JSONB NOT NULL,
    budget_max_yen      INTEGER NOT NULL,
    desired_date_from   TIMESTAMPTZ NOT NULL,
    desired_date_to     TIMESTAMPTZ NOT NULL,
    area_prefecture     VARCHAR(20) NOT NULL,
    area_city           VARCHAR(50),
    status              VARCHAR(20) NOT NULL DEFAULT 'open',
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_requests_user_id ON booking_requests(user_id);
CREATE INDEX idx_booking_requests_status ON booking_requests(status);
CREATE INDEX idx_booking_requests_area_prefecture ON booking_requests(area_prefecture);
CREATE INDEX idx_booking_requests_expires_at ON booking_requests(expires_at);

-- ----------------------------------------------------------------
-- bids
-- ----------------------------------------------------------------
CREATE TABLE bids (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_request_id      UUID NOT NULL REFERENCES booking_requests(id),
    salon_id                UUID NOT NULL REFERENCES salons(id),
    price_yen               INTEGER NOT NULL,
    includes_removal        BOOLEAN NOT NULL DEFAULT false,
    removal_fee_yen         INTEGER NOT NULL DEFAULT 0,
    available_slot_at       TIMESTAMPTZ NOT NULL,
    dynamic_discount_reason VARCHAR(100),
    message                 TEXT,
    rebid_count             INTEGER NOT NULL DEFAULT 0,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at              TIMESTAMPTZ NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(booking_request_id, salon_id),
    CHECK (rebid_count <= 1)
);

CREATE INDEX idx_bids_booking_request_id ON bids(booking_request_id);
CREATE INDEX idx_bids_salon_id ON bids(salon_id);
CREATE INDEX idx_bids_status ON bids(status);

CREATE TRIGGER trg_bids_updated_at
    BEFORE UPDATE ON bids
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ----------------------------------------------------------------
-- bookings
-- ----------------------------------------------------------------
CREATE TABLE bookings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_request_id  UUID NOT NULL UNIQUE REFERENCES booking_requests(id),
    bid_id              UUID NOT NULL UNIQUE REFERENCES bids(id),
    user_id             UUID NOT NULL REFERENCES users(id),
    salon_id            UUID NOT NULL REFERENCES salons(id),
    scheduled_at        TIMESTAMPTZ NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    confirmed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    cancellation_reason TEXT
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_salon_id ON bookings(salon_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_scheduled_at ON bookings(scheduled_at);

-- ----------------------------------------------------------------
-- payments
-- ----------------------------------------------------------------
CREATE TABLE payments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id                  UUID NOT NULL UNIQUE REFERENCES bookings(id),
    total_amount_yen            INTEGER NOT NULL,
    platform_fee_yen            INTEGER NOT NULL,
    salon_payout_yen            INTEGER NOT NULL,
    design_royalty_total_yen    INTEGER NOT NULL,
    stripe_payment_intent_id    VARCHAR(255) UNIQUE,
    stripe_charge_id            VARCHAR(255),
    payment_method              VARCHAR(50),
    status                      VARCHAR(20) NOT NULL DEFAULT 'authorized',
    authorized_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    captured_at                 TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_payments_amount_sum
        CHECK (total_amount_yen = salon_payout_yen + platform_fee_yen + design_royalty_total_yen)
);

CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(status);

-- ----------------------------------------------------------------
-- royalty_distributions
-- ----------------------------------------------------------------
CREATE TABLE royalty_distributions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      UUID NOT NULL REFERENCES payments(id),
    design_ip_id    UUID NOT NULL REFERENCES design_ips(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    amount_yen      INTEGER NOT NULL,
    share_percent   NUMERIC(5,2) NOT NULL,
    depth_level     INTEGER NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    credited_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(payment_id, user_id)
);

CREATE INDEX idx_royalty_distributions_payment_id ON royalty_distributions(payment_id);
CREATE INDEX idx_royalty_distributions_user_id ON royalty_distributions(user_id);
CREATE INDEX idx_royalty_distributions_status ON royalty_distributions(status);

-- ----------------------------------------------------------------
-- reviews
-- ----------------------------------------------------------------
CREATE TABLE reviews (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id              UUID NOT NULL UNIQUE REFERENCES bookings(id),
    user_id                 UUID NOT NULL REFERENCES users(id),
    salon_id                UUID NOT NULL REFERENCES salons(id),
    design_ip_id            UUID NOT NULL REFERENCES design_ips(id),
    reproduction_score      INTEGER NOT NULL CHECK (reproduction_score BETWEEN 1 AND 5),
    overall_score           INTEGER NOT NULL CHECK (overall_score BETWEEN 1 AND 5),
    comment                 TEXT,
    before_photo_url        TEXT,
    after_photo_url         TEXT,
    ai_reproduction_score   NUMERIC(3,2),
    ai_analysis_status      VARCHAR(20) DEFAULT 'pending',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_salon_id ON reviews(salon_id);
CREATE INDEX idx_reviews_design_ip_id ON reviews(design_ip_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);

-- ----------------------------------------------------------------
-- notifications
-- ----------------------------------------------------------------
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    type        VARCHAR(50) NOT NULL,
    title       VARCHAR(200) NOT NULL,
    body        TEXT,
    payload     JSONB NOT NULL DEFAULT '{}',
    is_read     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- 保持ポリシー: 既読かつ90日超のレコードをバッチ削除
);

CREATE INDEX idx_notifications_user_id_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
