# データベーススキーマ設計書（PostgreSQL 16）

## 拡張機能
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector; -- pgvector (512次元ベクトル)
```

---

## テーブル一覧

| テーブル名 | 概要 |
|-----------|------|
| users | ユーザー情報（consumer / salon_owner / admin）|
| nail_profiles | 爪カルテ（ユーザーごとの爪質データ）|
| ar_sessions | AR試着セッション（爪データ自動計測結果）|
| materials | 3D素材マスタ（パーツ・ベース・アート柄）|
| design_ips | デザインIP本体 |
| design_royalty_nodes | IPロイヤリティ分配ツリー |
| salons | サロン情報 |
| booking_requests | 予約リクエスト（逆オークション出品）|
| bids | 入札（サロンからの一発提示）|
| bookings | 確定予約 |
| payments | 決済記録 |
| royalty_distributions | IP利用料分配記録 |
| reviews | レビュー・再現度評価 |
| notifications | 通知履歴 |

---

## DDL

### users
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid    VARCHAR(128) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    gender          VARCHAR(30),      -- 'male', 'female', 'nonbinary', 'prefer_not_to_say', or free text
    role            VARCHAR(20) NOT NULL DEFAULT 'consumer', -- 'consumer', 'salon_owner', 'admin'
    avatar_url      TEXT,
    bio             TEXT,
    lifestyle_tags  TEXT[] DEFAULT '{}', -- ['water_work', 'sports', 'keyboard_heavy', 'outdoor']
    point_balance   INTEGER NOT NULL DEFAULT 0, -- IPロイヤリティ獲得ポイント残高
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_role ON users(role);
```

### nail_profiles
```sql
CREATE TABLE nail_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    avg_nail_length_mm  NUMERIC(4,1),
    nail_shape          VARCHAR(30), -- 'round','square','oval','almond','stiletto','coffin'
    gel_lift_tendency   VARCHAR(20), -- 'low','medium','high'
    allergy_notes       TEXT,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);
```

### ar_sessions
```sql
-- AR試着時に自動計測された爪データのスナップショット
-- このデータが予約リクエストに自動添付され、サロンの入札精度を上げるMoatの核心
CREATE TABLE ar_sessions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    design_ip_id                UUID,     -- 試着したデザイン（NULL=デザイン未選択）
    detected_nail_length_mm     NUMERIC(4,1),
    has_existing_gel            BOOLEAN,
    detected_nail_shape         VARCHAR(30),
    estimated_treatment_min     INTEGER,  -- バックエンド算出の推定施術時間
    estimated_gel_amount_ml     NUMERIC(5,3), -- バックエンド算出の推定ジェル量
    hand_snapshot_url           TEXT,     -- S3 URL（署名付きURL発行。90日保持）
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ar_sessions_user_id ON ar_sessions(user_id);
CREATE INDEX idx_ar_sessions_created_at ON ar_sessions(created_at DESC);
```

### materials
```sql
CREATE TABLE materials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    category        VARCHAR(50) NOT NULL,  -- 'base', 'part', 'art', 'glitter', 'foil'
    texture_type    VARCHAR(50),           -- 'matte', 'mirror', 'magnet', '3d_puffy', 'normal'
    thumbnail_url   TEXT NOT NULL,
    model_3d_url    TEXT,                  -- Three.js用GLTFモデルのS3 URL
    texture_params  JSONB NOT NULL DEFAULT '{}',
    -- 例: {"shininess": 0.8, "bumpScale": 0.5, "roughness": 0.2}
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_materials_category ON materials(category);
CREATE INDEX idx_materials_is_active ON materials(is_active);
```

### design_ips
```sql
CREATE TABLE design_ips (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id          UUID NOT NULL REFERENCES users(id),
    parent_ip_id        UUID REFERENCES design_ips(id) ON DELETE SET NULL, -- フォーク元（NULLなら独立IP）
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    preview_image_url   TEXT NOT NULL,  -- S3 URL（エディターのレンダリング結果）

    -- 3Dデザイン定義データ（構造化JSON）
    design_data         JSONB NOT NULL,
    -- 例:
    -- {
    --   "fingers": {
    --     "thumb":  [{"material_id": "...", "position": {"x":0,"y":0}, "rotation": 0, "texture_overrides": {"shininess": 0.9}}],
    --     "index":  [...],
    --     "middle": [...],
    --     "ring":   [...],
    --     "pinky":  [...]
    --   }
    -- }

    -- pgvector: SageMakerで生成した512次元埋め込みベクトル（類似度検索用）
    similarity_vector   vector(512),
    similarity_hash     VARCHAR(64),    -- design_dataの高速重複チェック用ハッシュ

    fork_depth          INTEGER NOT NULL DEFAULT 0,  -- 0=独立IP, 1=1段派生, 2=2段派生...
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending'(審査中), 'active'(公開), 'rejected'(拒否), 'archived'(非公開)

    is_public           BOOLEAN NOT NULL DEFAULT true,
    gender_tag          VARCHAR(20) NOT NULL DEFAULT 'neutral', -- 'feminine','masculine','neutral'
    style_tags          TEXT[] DEFAULT '{}',

    usage_count         INTEGER NOT NULL DEFAULT 0,  -- このIPで施術された回数（denormalized）
    total_royalty_yen   INTEGER NOT NULL DEFAULT 0,  -- 累計ロイヤリティ収益（denormalized）

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 類似度ベクトル検索用インデックス（HNSW: pgvector v0.5+ 推奨。コサイン類似度で検索するため vector_cosine_ops を使用）
-- ※ IVFFlat + vector_l2_ops は誤り。L2距離とコサイン類似度は異なる演算で、クエリと不一致だとインデックスが効かず全件スキャンになる
-- ※ IVFFlat は初期データが0件の場合 lists パラメータが無効になるため、HNSW を採用
CREATE INDEX idx_design_ips_similarity_vector ON design_ips USING hnsw (similarity_vector vector_cosine_ops);
CREATE INDEX idx_design_ips_creator_id ON design_ips(creator_id);
CREATE INDEX idx_design_ips_status ON design_ips(status);
CREATE INDEX idx_design_ips_gender_tag ON design_ips(gender_tag);
CREATE INDEX idx_design_ips_parent_ip_id ON design_ips(parent_ip_id);
CREATE INDEX idx_design_ips_usage_count ON design_ips(usage_count DESC);
```

### design_royalty_nodes
```sql
-- フォーク派生時のロイヤリティ分配ツリー
-- design_ip_idが施術されるたびに、このテーブルの全ノードに分配が実行される
-- 多段フォーク深さ上限: 3段（要確認→open_questions参照）
CREATE TABLE design_royalty_nodes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    design_ip_id    UUID NOT NULL REFERENCES design_ips(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    share_percent   NUMERIC(5,2) NOT NULL, -- 例: 70.00, 21.00, 9.00
    depth_level     INTEGER NOT NULL,      -- 0=当該デザイナー, 1=親, 2=祖父
    UNIQUE(design_ip_id, user_id)
);

CREATE INDEX idx_royalty_nodes_design_ip_id ON design_royalty_nodes(design_ip_id);
```

### salons
```sql
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
    -- 例: {"mon": {"open": "10:00", "close": "20:00", "closed": false}, ...}

    avg_reproduction_score  NUMERIC(3,2) NOT NULL DEFAULT 0.00, -- 0.00-5.00（レビューから集計）
    skill_badge_tags        TEXT[] DEFAULT '{}',
    -- 例: ['3d_parts', 'mirror_gel', 'mens_nail', 'short_nail']
    portfolio_image_urls    TEXT[] DEFAULT '{}',

    verification_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending', 'verified', 'rejected', 'suspended'
    is_active               BOOLEAN NOT NULL DEFAULT false, -- 審査通過後にtrueへ

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_salons_prefecture ON salons(prefecture);
CREATE INDEX idx_salons_verification_status ON salons(verification_status);
CREATE INDEX idx_salons_avg_reproduction_score ON salons(avg_reproduction_score DESC);
-- 位置情報検索用（将来的にPostGIS導入を検討）
CREATE INDEX idx_salons_lat_lng ON salons(lat, lng);
```

### booking_requests
```sql
-- 逆オークションの「出品」エンティティ
-- ユーザーが条件を投稿し、サロン側が入札する
CREATE TABLE booking_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    design_ip_id        UUID NOT NULL REFERENCES design_ips(id),
    ar_session_id       UUID REFERENCES ar_sessions(id), -- AR試着データ（NULLなら手動入力）

    -- ARデータスナップショット（ar_sessionが削除されても参照可能なように複製保持）
    nail_data_snapshot  JSONB NOT NULL,
    -- 例: {"length_mm": 8.5, "has_existing_gel": true, "shape": "oval",
    --       "estimated_treatment_min": 90, "estimated_gel_amount_ml": 2.3}

    budget_max_yen      INTEGER NOT NULL,
    desired_date_from   TIMESTAMPTZ NOT NULL,
    desired_date_to     TIMESTAMPTZ NOT NULL,
    area_prefecture     VARCHAR(20) NOT NULL,
    area_city           VARCHAR(50),

    status              VARCHAR(20) NOT NULL DEFAULT 'open',
    -- 'open'(受付中), 'bidding'(入札あり), 'confirmed'(予約確定), 'cancelled', 'expired'

    expires_at          TIMESTAMPTZ NOT NULL, -- 入札受付締切（作成から24時間後）
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_requests_user_id ON booking_requests(user_id);
CREATE INDEX idx_booking_requests_status ON booking_requests(status);
CREATE INDEX idx_booking_requests_area_prefecture ON booking_requests(area_prefecture);
CREATE INDEX idx_booking_requests_expires_at ON booking_requests(expires_at);
```

### bids
```sql
-- サロンからの一発入札
-- 1サロン1リクエストにつき1入札のみ（UNIQUE制約）
-- 再入札は価格引き下げのみ許可（要UPDATE。INSERTは不可）
CREATE TABLE bids (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_request_id      UUID NOT NULL REFERENCES booking_requests(id),
    salon_id                UUID NOT NULL REFERENCES salons(id),

    price_yen               INTEGER NOT NULL,
    includes_removal        BOOLEAN NOT NULL DEFAULT false, -- ジェルオフ込みか
    removal_fee_yen         INTEGER NOT NULL DEFAULT 0,

    available_slot_at       TIMESTAMPTZ NOT NULL,         -- 提示可能な施術日時
    dynamic_discount_reason VARCHAR(100),                 -- '直前空き枠30分前', '平日午前限定' 等
    message                 TEXT,                         -- サロンからの一言（任意）

    rebid_count             INTEGER NOT NULL DEFAULT 0,   -- 再入札回数（最大1回まで）
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending', 'accepted', 'rejected', 'cancelled_by_salon', 'expired'

    expires_at              TIMESTAMPTZ NOT NULL,         -- 希望日時の12時間前
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(booking_request_id, salon_id),
    CHECK (rebid_count <= 1)
);

CREATE INDEX idx_bids_booking_request_id ON bids(booking_request_id);
CREATE INDEX idx_bids_salon_id ON bids(salon_id);
CREATE INDEX idx_bids_status ON bids(status);
```

### bookings
```sql
CREATE TABLE bookings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_request_id  UUID NOT NULL UNIQUE REFERENCES booking_requests(id),
    bid_id              UUID NOT NULL UNIQUE REFERENCES bids(id),
    user_id             UUID NOT NULL REFERENCES users(id),
    salon_id            UUID NOT NULL REFERENCES salons(id),
    scheduled_at        TIMESTAMPTZ NOT NULL,

    status              VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    -- 'confirmed', 'completed', 'cancelled_by_user', 'cancelled_by_salon', 'no_show'

    confirmed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    cancellation_reason TEXT
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_salon_id ON bookings(salon_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_scheduled_at ON bookings(scheduled_at);
```

### payments
```sql
-- 決済記録
-- エスクロー方式: booking確定時にオーソリ → 施術完了後にキャプチャ
CREATE TABLE payments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id                  UUID NOT NULL UNIQUE REFERENCES bookings(id),

    total_amount_yen            INTEGER NOT NULL,
    platform_fee_yen            INTEGER NOT NULL,       -- プラットフォーム手数料（15%）
    salon_payout_yen            INTEGER NOT NULL,       -- サロン受取額
    design_royalty_total_yen    INTEGER NOT NULL,       -- IP利用料総額（5%）

    stripe_payment_intent_id    VARCHAR(255) UNIQUE,
    stripe_charge_id            VARCHAR(255),
    payment_method              VARCHAR(50),            -- 'card', 'apple_pay', 'google_pay'

    status                      VARCHAR(20) NOT NULL DEFAULT 'authorized',
    -- 'authorized'(オーソリ済み), 'captured'(キャプチャ済み), 'refunded', 'failed'

    authorized_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    captured_at                 TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_payments_amount_sum
        CHECK (total_amount_yen = salon_payout_yen + platform_fee_yen + design_royalty_total_yen)
);

CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(status);
```

### royalty_distributions
```sql
-- IP利用料の分配記録（payments.captured後に非同期生成）
CREATE TABLE royalty_distributions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      UUID NOT NULL REFERENCES payments(id),
    design_ip_id    UUID NOT NULL REFERENCES design_ips(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    amount_yen      INTEGER NOT NULL,
    share_percent   NUMERIC(5,2) NOT NULL,
    depth_level     INTEGER NOT NULL,   -- 0=直接デザイナー, 1=親IP, 2=祖父IP

    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending', 'credited'(ポイント付与済み), 'paid'(現金振込済み), 'cancelled'

    credited_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_royalty_distributions_payment_id ON royalty_distributions(payment_id);
CREATE INDEX idx_royalty_distributions_user_id ON royalty_distributions(user_id);
CREATE INDEX idx_royalty_distributions_status ON royalty_distributions(status);
```

### reviews
```sql
CREATE TABLE reviews (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id              UUID NOT NULL UNIQUE REFERENCES bookings(id),
    user_id                 UUID NOT NULL REFERENCES users(id),
    salon_id                UUID NOT NULL REFERENCES salons(id),
    design_ip_id            UUID NOT NULL REFERENCES design_ips(id),

    reproduction_score      INTEGER NOT NULL CHECK (reproduction_score BETWEEN 1 AND 5),
    overall_score           INTEGER NOT NULL CHECK (overall_score BETWEEN 1 AND 5),
    comment                 TEXT,

    before_photo_url        TEXT,   -- AR試着スクリーンショット or 施術前写真
    after_photo_url         TEXT,   -- 施術後写真

    -- AI解析スコア（SageMakerによるbefore/after類似度）
    ai_reproduction_score   NUMERIC(3,2),   -- 0.00-1.00（1が完全一致）
    ai_analysis_status      VARCHAR(20) DEFAULT 'pending',  -- 'pending','completed','failed'

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_salon_id ON reviews(salon_id);
CREATE INDEX idx_reviews_design_ip_id ON reviews(design_ip_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
```

### notifications
```sql
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    type        VARCHAR(50) NOT NULL,
    -- 'new_bid', 'bid_accepted', 'bid_rejected', 'booking_confirmed',
    -- 'booking_completed', 'royalty_earned', 'review_request', 'ip_approved', 'ip_rejected'
    title       VARCHAR(200) NOT NULL,
    body        TEXT,
    payload     JSONB NOT NULL DEFAULT '{}',  -- 画面遷移用パラメータ
    is_read     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- 保持ポリシー: 既読かつ90日超のレコードをバッチ削除（pg_cron or Lambda で定期実行）
);

CREATE INDEX idx_notifications_user_id_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
```

---

## updated_at 自動更新トリガー

PostgreSQL の `DEFAULT NOW()` は INSERT 時のみ適用される。UPDATE 時の自動更新には以下のトリガーが必要。

```sql
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_salons_updated_at
    BEFORE UPDATE ON salons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_design_ips_updated_at
    BEFORE UPDATE ON design_ips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_bids_updated_at
    BEFORE UPDATE ON bids
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_nail_profiles_updated_at
    BEFORE UPDATE ON nail_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

## 主要な制約・ルール

| ルール | 実装箇所 |
|--------|---------|
| 同一ユーザーが同一IPを施術した場合もロイヤリティ発生（自作自演防止は別途検討）| アプリロジック |
| bids: 同一サロンは同一リクエストに1回のみ | UNIQUE(booking_request_id, salon_id) |
| bids: 再入札は価格引き下げのみ・1回まで | CHECK (rebid_count <= 1) + アプリロジック |
| booking_requests: 1ユーザーが同時に持てるopenリクエストは3件まで | アプリロジック |
| design_royalty_nodes: fork_depth上限は3 | アプリロジック |
| payments: キャプチャは必ず施術完了（bookings.status='completed'）後に実行 | アプリロジック |
| payments: total = salon + platform_fee + royalty の整合性 | CHECK制約 (chk_payments_amount_sum) |

---

## マイグレーション管理
- ツール: `golang-migrate`
- ファイル形式: `migrations/000001_init.up.sql` / `000001_init.down.sql`
- 環境: dev / staging / prod それぞれに独立したDB
