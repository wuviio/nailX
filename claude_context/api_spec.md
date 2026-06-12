# API仕様書（REST API エンドポイント一覧）

## 共通ルール

| 項目 | 内容 |
|------|------|
| Base URL | `/api/v1` |
| 認証ヘッダー | `Authorization: Bearer <Firebase ID Token>` |
| Content-Type | `application/json` |
| エラー形式 | `{"error": {"code": "ERROR_CODE", "message": "..."}}` |
| ページネーション | `?limit=20&cursor=<opaque_cursor>` （カーソルベース）。カーソル形式: `base64url(created_at_unix_ms + ":" + id)`|
| 日時形式 | RFC3339 (例: `2026-06-03T10:00:00+09:00`) |

### 認可ルール
- `[Public]`: 認証不要
- `[Consumer]`: consumerロール以上
- `[Salon]`: salon_ownerロール
- `[Admin]`: adminロール

---

## 1. 認証・ユーザー

### `POST /auth/register`
Firebase ID Tokenを検証し、DBにユーザーレコードを作成。
```json
Request: { "firebase_token": "...", "display_name": "田中 美優", "gender": "female" }
Response 201: { "user": { "id": "uuid", "display_name": "...", "role": "consumer" } }
```

### `GET /users/me` [Consumer]
自分のプロフィール取得。
```json
Response 200: {
  "user": { "id": "uuid", "display_name": "...", "gender": "...", "point_balance": 1200 },
  "nail_profile": { "avg_nail_length_mm": 8.5, "nail_shape": "oval", ... }
}
```

### `PATCH /users/me` [Consumer]
プロフィール更新（display_name / avatar_url / lifestyle_tags等）。

### `GET /users/:id` [Public]
他ユーザーの公開プロフィール・作成IPデザイン一覧取得。

---

## 2. 爪カルテ

### `PUT /users/me/nail-profile` [Consumer]
爪カルテの作成または更新（UPSERT）。
```json
Request: {
  "avg_nail_length_mm": 8.5,
  "nail_shape": "oval",
  "gel_lift_tendency": "medium",
  "allergy_notes": "ハードジェルアレルギーあり"
}
Response 200: { "nail_profile": { ... } }
```

---

## 3. AR試着

### `POST /ar/sessions` [Consumer]
AR試着セッションの保存。Flutter(DeepAR)側で計測した爪データを送信。

> **S3アップロードフロー**: クライアントは先に `POST /media/presigned-url` (`purpose: "ar_snapshot"`) で署名付きURLを取得し、S3に直接アップロードする。その際レスポンスで返却された `file_url` を本エンドポイントに渡す。クライアントが任意の `s3://` URIを直接送信することは禁止（サーバー側でバケット・プレフィックスのバリデーションを実施）。

```json
Request: {
  "design_ip_id": "uuid",
  "detected_nail_length_mm": 8.2,
  "has_existing_gel": true,
  "detected_nail_shape": "oval",
  "hand_snapshot_url": "https://cdn.nailx.jp/ar-snapshots/uuid.jpg"
  // ← POST /media/presigned-url のレスポンス file_url（CDN経由URL）を使用
}
Response 201: {
  "ar_session": {
    "id": "uuid",
    "estimated_treatment_min": 90,
    "estimated_gel_amount_ml": 2.3
  }
}
```

### `GET /ar/sessions/:id` [Consumer]
AR試着セッション詳細取得（自分のセッションのみ）。

---

## 4. 素材マスタ

### `GET /materials` [Public]
素材一覧取得。カテゴリフィルター可。
```
Query: ?category=part&texture_type=3d_puffy&limit=50
Response 200: { "materials": [...], "next_cursor": "..." }
```

### `POST /materials` [Admin]
新規素材追加。

### `PATCH /materials/:id` [Admin]
素材の有効/無効切替・情報更新。

---

## 5. デザインIP

### `POST /designs` [Consumer]
デザインIP登録リクエスト（非同期処理）。

> `preview_image_url` は `POST /media/presigned-url` (`purpose: "design_preview"`) で取得した `file_url` を使用。`s3://` URIの直接送信は禁止。

```json
Request: {
  "title": "シンプルミラーメンズネイル",
  "description": "...",
  "preview_image_url": "https://cdn.nailx.jp/designs/preview/uuid.jpg",
  "design_data": {
    "fingers": {
      "thumb":  [{"material_id": "uuid", "position": {"x": 0, "y": 0}, "rotation": 0, "texture_overrides": {}}],
      "index":  [...],
      "middle": [...],
      "ring":   [...],
      "pinky":  [...]
    }
  },
  "parent_ip_id": null,
  "gender_tag": "masculine",
  "style_tags": ["mirror", "simple", "mens"],
  "is_public": true
}
Response 202: { "design_ip_id": "uuid", "status": "pending", "job_id": "..." }
```

### `GET /designs` [Public]
デザインフィード一覧。
```
Query: ?gender_tag=neutral&style_tags=mirror,3d&status=active&sort=usage_count&limit=20&cursor=...
Response 200: { "designs": [...], "next_cursor": "..." }
```

### `GET /designs/:id` [Public]
デザインIP詳細。フォーク系譜・ロイヤリティノード情報含む。
```json
Response 200: {
  "design": { "id": "uuid", "title": "...", "usage_count": 342, ... },
  "creator": { "id": "uuid", "display_name": "..." },
  "parent_ip": { "id": "uuid", "title": "..." },
  "royalty_nodes": [
    { "user_id": "uuid", "share_percent": 70, "depth_level": 0 },
    { "user_id": "uuid", "share_percent": 30, "depth_level": 1 }
  ]
}
```

### `GET /designs/similarity-check` [Consumer]
デザイン登録前の類似度プレチェック（エディター内でのリアルタイム確認用）。
```
Query: ?job_id=... （POST /designs の job_id）
Response 200: { "status": "completed", "result": "fork", "similar_ip_id": "uuid", "score": 0.82 }
```

### `GET /users/:id/designs` [Public]
特定ユーザーが作成したデザイン一覧。

### `PATCH /designs/:id` [Consumer]
タイトル・説明・公開設定の更新（design_dataは変更不可。変更する場合は新規フォーク）。

---

## 6. サロン

### `POST /salons` [Consumer]
サロン登録申請（Consumer → salon_ownerへのロール変更は審査後）。
```json
Request: {
  "name": "Nail Atelier KIRA",
  "address": "東京都渋谷区...",
  "prefecture": "東京",
  "city": "渋谷区",
  "lat": 35.6580,
  "lng": 139.7016,
  "phone": "03-xxxx-xxxx",
  "business_hours": { "mon": {"open": "10:00", "close": "20:00", "closed": false}, ... }
}
Response 202: { "salon_id": "uuid", "verification_status": "pending" }
```

### `GET /salons` [Public]
サロン一覧。エリア・バッジタグフィルター可。
```
Query: ?prefecture=東京&skill_badge_tags=mens_nail,3d_parts&sort=avg_reproduction_score
```

### `GET /salons/:id` [Public]
サロン詳細（ポートフォリオ・バッジ・スコア含む）。

### `PATCH /salons/:id` [Salon]
サロン情報更新（自サロンのみ）。

### `POST /salons/:id/portfolio` [Salon]
施術写真の追加（S3アップロード済みURLを登録）。

---

## 7. 逆オークション（予約リクエスト）

### `POST /auctions/requests` [Consumer]
予約リクエスト（逆オークション出品）作成。
```json
Request: {
  "design_ip_id": "uuid",
  "ar_session_id": "uuid",
  "nail_data_snapshot": {
    "length_mm": 8.2,
    "has_existing_gel": true,
    "shape": "oval",
    "estimated_treatment_min": 90,
    "estimated_gel_amount_ml": 2.3
  },
  "budget_max_yen": 8000,
  "desired_date_from": "2026-06-10T10:00:00+09:00",
  "desired_date_to": "2026-06-12T20:00:00+09:00",
  "area_prefecture": "東京",
  "area_city": "渋谷区"
}
Response 201: { "booking_request": { "id": "uuid", "status": "open", "expires_at": "..." } }
```

### `GET /auctions/requests/:id` [Consumer]
自分の予約リクエスト詳細 + 入札一覧取得。

### `DELETE /auctions/requests/:id` [Consumer]
予約リクエストのキャンセル（status=open/biddingのみ可）。

### `GET /auctions/requests` [Salon]
サロン向け: 自サロンのエリア・スキルにマッチしたオープンリクエスト一覧。
```
Query: ?sort=expires_at&limit=20&cursor=...
```

---

## 8. 入札

### `POST /auctions/requests/:request_id/bids` [Salon]
入札（一発提示）。
```json
Request: {
  "price_yen": 6500,
  "includes_removal": true,
  "removal_fee_yen": 0,
  "available_slot_at": "2026-06-10T14:00:00+09:00",
  "dynamic_discount_reason": "直前空き枠割引",
  "message": "3Dパーツの再現度には自信があります"
}
Response 201: { "bid": { "id": "uuid", "status": "pending", "expires_at": "..." } }
```
- 同一サロンの再入札: 同エンドポイントへPATCHで価格引き下げのみ可

### `PATCH /auctions/requests/:request_id/bids/:bid_id` [Salon]
入札価格の修正（引き下げのみ。1回まで）。

### `GET /auctions/requests/:request_id/bids` [Consumer]
自分のリクエストへの入札一覧取得。
```
Query: ?sort=price_asc|score_desc|distance
```

### `GET /auctions/requests/:request_id/bids/stream` [Consumer]
SSEエンドポイント。入札リクエストへの新着入札をリアルタイム受信。
```
Response: text/event-stream
event: new_bid
data: { "bid": { "id": "uuid", "price_yen": 6500, "salon": {...} } }
```

---

## 9. 予約確定・決済

### `POST /bookings` [Consumer]
入札選択 → 予約確定。Stripe Payment Intent作成。
```json
Request: {
  "bid_id": "uuid",
  "payment_method_id": "pm_stripe_..."
}
Response 201: {
  "booking": { "id": "uuid", "scheduled_at": "...", "status": "confirmed" },
  "payment": { "id": "uuid", "total_amount_yen": 6500, "status": "authorized" },
  "stripe_client_secret": "pi_..._secret_..."
}
```

### `GET /bookings` [Consumer]
自分の予約一覧。
```
Query: ?status=confirmed|completed|cancelled
```

### `GET /bookings/:id` [Consumer/Salon]
予約詳細。

### `POST /bookings/:id/complete` [Salon]
施術完了報告（→ 決済キャプチャ → ロイヤリティ分配 → レビュー依頼通知）。

### `POST /bookings/:id/cancel` [Consumer/Salon]
予約キャンセル申請。
```json
Request: { "reason": "体調不良のため" }
```

---

## 10. レビュー

### `POST /reviews` [Consumer]
レビュー投稿（施術完了から7日以内）。

> **S3アップロードフロー**: `before_photo_url` / `after_photo_url` は `POST /media/presigned-url` (`purpose: "review_photo"`) で取得した `file_url` を使用。`s3://` URIの直接送信は禁止。

```json
Request: {
  "booking_id": "uuid",
  "reproduction_score": 5,
  "overall_score": 5,
  "comment": "デザイン通りに再現してくれました！",
  "before_photo_url": "https://cdn.nailx.jp/reviews/before/uuid.jpg",
  "after_photo_url": "https://cdn.nailx.jp/reviews/after/uuid.jpg"
}
Response 201: { "review": { "id": "uuid" } }
```

### `GET /salons/:id/reviews` [Public]
サロンのレビュー一覧。

---

## 11. 通知

### `GET /notifications` [Consumer/Salon]
通知一覧。
```
Query: ?is_read=false&limit=20&cursor=...
```

### `PATCH /notifications/:id/read` [Consumer/Salon]
通知を既読に。

### `POST /notifications/fcm-token` [Consumer/Salon]
FCMトークン登録・更新。
```json
Request: { "fcm_token": "...", "device_platform": "ios" }
```

---

## 12. Admin

### `GET /admin/salons` [Admin]
審査待ちサロン一覧。

### `PATCH /admin/salons/:id/verify` [Admin]
サロン審査承認/却下。
```json
Request: { "action": "approve" | "reject", "reason": "..." }
```

### `GET /admin/designs/flagged` [Admin]
盗作フラグが立ったデザイン一覧。

### `PATCH /admin/designs/:id/moderate` [Admin]
IP審査の手動上書き。
```json
Request: { "action": "approve" | "reject", "reason": "..." }
```

### `GET /admin/settings/similarity-threshold` [Admin]
類似度閾値の確認・更新。
```json
Response: { "fork_threshold": 0.75, "reject_threshold": 0.92 }
PATCH: { "fork_threshold": 0.78 }
```

---

## S3署名付きURL発行

### `POST /media/presigned-url` [Consumer/Salon]
クライアントが直接S3にアップロードするための署名付きURLを発行。
`purpose` ごとに書き込みバケット・プレフィックスが固定される（任意パスへの書き込みを防ぐ）。

```json
Request: {
  "file_type": "image/jpeg",
  "purpose": "ar_snapshot" | "portfolio" | "review_photo" | "design_preview"
}
Response 200: {
  "upload_url": "https://nailx-media.s3.ap-northeast-1.amazonaws.com/ar-snapshots/uuid?X-Amz-...",
  "file_url": "https://cdn.nailx.jp/ar-snapshots/uuid.jpg"
  // file_url は CDN 経由 URL。各エンドポイントの *_url フィールドにこちらを渡す
}
```

> **セキュリティ注記**: サーバーは `file_url` 受信時にバケット名・プレフィックスが期待パターンと一致するかを検証し、他バケットや任意パスへの参照を拒否する。

---

## エラーコード一覧

| HTTP | Error Code | 意味 |
|------|-----------|------|
| 400 | INVALID_REQUEST | リクエストパラメータ不正 |
| 401 | UNAUTHORIZED | 認証トークン無効 |
| 403 | FORBIDDEN | 権限不足 |
| 404 | NOT_FOUND | リソースが存在しない |
| 409 | ALREADY_BID | 既に入札済み |
| 409 | IP_REJECTED | 盗作判定でIP登録拒否 |
| 409 | REQUEST_EXPIRED | リクエスト有効期限切れ |
| 422 | BUDGET_EXCEEDED | 入札価格が予算上限超過 |
| 429 | RATE_LIMITED | レートリミット超過 |
| 500 | INTERNAL_ERROR | サーバーエラー |
