# 実装タスク一覧（Implementation Tasks）

## 進め方の原則
- 実装はユーザーの承認後に開始
- 各フェーズ完了後にレビュー・動作確認を挟む
- open_questions.md の 🔴 事項はPhase 0完了前に全て確定させること

---

## Phase 0: 基盤構築 ✅ 完了

### 0-1. プロジェクト構造セットアップ ✅
```
nailX/
├── backend/          # Go API サーバー
│   ├── cmd/server/
│   ├── internal/
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   └── middleware/
│   ├── migrations/
│   └── sqlc.yaml
├── frontend/         # Flutter アプリ
│   ├── lib/
│   │   ├── features/
│   │   ├── core/
│   │   └── shared/
│   └── pubspec.yaml
├── infra/            # Terraform / K8s manifests（Phase 0-5 は未実装）
│   ├── terraform/
│   └── k8s/
└── claude_context/   # このディレクトリ（仕様書群）
```

### 0-2. DB マイグレーション基盤 ✅
- [x] `golang-migrate` セットアップ
- [x] 全テーブルDDL（db_schema.md参照）の `001_init.up.sql` 作成（HNSW インデックス + updated_at トリガー含む）
- [x] pgvector拡張の有効化
- [x] ローカル開発用Docker Compose（PostgreSQL + pgvector + Redis）

### 0-3. Go バックエンド基盤 ✅
- [x] Echoフレームワーク セットアップ
- [x] sqlc.yaml 設定（Phase 2 以降でクエリ生成実施）
- [x] Firebase Admin SDK統合（JWT検証ミドルウェア）
- [x] 環境変数管理（DATABASE_URL / REDIS_URL / FIREBASE_CREDENTIALS_PATH / CDN_BASE_URL / S3_BUCKET / SQS_SIMILARITY_QUEUE_URL）
- [x] ヘルスチェックエンドポイント `/health`
- [x] 構造化ログ（slog JSON形式）
- [x] エラーハンドリング共通化（response.go + domain errors）
- [x] 全サービス・リポジトリの実配線（cmd/server/main.go）

### 0-4. Flutter 基盤 ✅
- [x] Riverpod 2.x セットアップ
- [x] Firebase Auth（Google/Apple）統合（firebase_options.dart スタブ）
- [x] Dio + APIクライアント基盤（interceptor: Firebase Token自動付与）
- [x] go_router によるルーティング設計（全画面配線済み）
- [x] デザインシステム（カラー・タイポグラフィ: Material3 + ピンク主色）

### 0-5. インフラ基盤 ❌ 未着手
- [ ] AWS EKS クラスター構築（Terraform）
- [ ] RDS PostgreSQL Multi-AZ
- [ ] S3バケット（メディア用・ライフサイクルポリシー設定）
- [ ] CloudFront CDN
- [ ] ECR（Dockerイメージレジストリ）
- [ ] **ElastiCache (Redis)** — SSE マルチPodブロードキャスト用 Pub/Sub
- [ ] CI/CD パイプライン（GitHub Actions: test → build → push → deploy）

> **VS Code Dev Container**: `.devcontainer/` 設定済み（Dockerfile / docker-compose.devcontainer.yml / devcontainer.json / post-create.sh）

---

## Phase 1: コアデータモデル & 基本 CRUD API ✅ 完了

### 1-1. ユーザー管理 ✅
- [x] `POST /auth/register` — Firebase Token検証 & ユーザー作成
- [x] `GET /users/me` — プロフィール取得（nail_profile JOIN）
- [x] `PATCH /users/me` — プロフィール更新
- [x] `GET /users/:id` — 公開プロフィール取得

### 1-2. 爪カルテ ✅
- [x] `PUT /users/me/nail-profile` — UPSERT

### 1-3. 素材マスタ ⚠️ 部分実装
- [x] `POST /materials` [Admin] — 素材追加（ハンドラー定義済み）
- [x] `PATCH /materials/:id` [Admin] — 有効/無効切替（ハンドラー定義済み）
- [ ] `GET /materials` — 一覧取得ルーティング（ TODO: main.go に追加が必要）
- [ ] S3への素材ファイルアップロードフロー（Phase 2 で整備）

### 1-4. サロン管理 ✅
- [x] `POST /salons` — 申請
- [x] `GET /salons` — 一覧（エリア・バッジフィルター）
- [x] `GET /salons/:id` — 詳細
- [x] `PATCH /salons/:id` [Salon] — 情報更新
- [x] `PATCH /admin/salons/:id/verify` [Admin] — 審査
- [x] `POST /salons/:id/portfolio` — 施術写真追加

---

## Phase 2: デザイン IP システム ⚠️ 部分完了

### 2-1. 3D エディター（Flutter） ❌ 未実装
- [ ] Three.js WebViewブリッジ実装（Dart ↔ JavaScript通信）
- [ ] パーツライブラリUI（カテゴリタブ・グリッド）
- [ ] ドラッグ&ドロップで5本指キャンバスへ配置
- [ ] 質感パラメータスライダーUI
- [ ] 指別設定 / 全指コピー機能
- [ ] デザイン確定時のpreview_image_url生成（Three.jsレンダリングをキャプチャ）
- [ ] design_dataのJSONシリアライズ & Go APIへの送信

> 現状: `DesignNewScreen` にタイトル・タグ設定UIはあるが、3Dキャンバスはプレースホルダー

### 2-2. デザイン IP 登録（Go バックエンド） ✅
- [x] `POST /designs` — 登録リクエスト受付（非同期: SQSへエンキュー）
- [x] `POST /media/presigned-url` — S3署名付きURL発行（CDNバリデーション付き）
- [x] SQS Worker: SageMaker呼び出し → embedding生成 → pgvector類似度検索（`similarity_worker.go`）
- [x] 類似度判定ロジック実装（`<=>` コサイン演算子 + HNSW インデックス）
  - 独立IP / フォークIP / 盗作判定 + `design_royalty_nodes` ツリー生成
- [x] `GET /designs/similarity-check` — 判定状況ポーリング
- [x] `GET /designs` — フィード一覧（pgvector・タグフィルター・ソート・カーソルページネーション）
- [x] `GET /designs/:id` — 詳細（royalty_nodes含む）

### 2-3. デザイン IP 画面（Flutter） ✅
- [x] デザインフィード画面（C-10）— 2列グリッド + フィルターチップ
- [x] デザイン詳細画面（C-12）— フォーク元・ロイヤリティ分配表示
- [x] デザイン作成画面（C-23 相当）— タイトル・タグ入力（3Dエディターはプレースホルダー）
- [x] マイデザイン一覧（C-55）— プロフィール画面に統合

---

## Phase 3: AR 試着 ⚠️ バックエンドのみ完了

### 3-1. DeepAR SDK 統合（Flutter） ❌ 未実装

> **⚠️ 注意: DeepAR 公式 Flutter プラグインは存在しない（2025年時点）**  
> iOS / Android ネイティブ SDK を Flutter の **Method Channel（Platform Channel）** でブリッジする実装が必要。  
> Phase 3 着手前に以下のいずれかを検証すること:
> - Option A: `flutter_arkit`（iOS）+ `ar_flutter_plugin`（Android）でプロトタイプ検証 → DeepAR の精度が必要な場合は Option B へ
> - Option B: iOS 側に DeepAR iOS SDK を組み込み、`FlutterMethodChannel` 経由で `startCamera` / `getNailData` 等を呼び出すカスタムブリッジを実装。Android 側も同様に DeepAR Android SDK + Method Channel。
> - **工数見積もり**: Option A は 1〜2週間、Option B は 3〜4週間（iOS + Android）

- [ ] **Phase 3 着手前**: DeepAR Flutter 統合方式の PoC（上記 Option A or B を選定）
- [ ] Method Channel ブリッジ実装（iOS: `AppDelegate` / Android: `MainActivity` に DeepAR SDK を初期化）
- [ ] カメラパーミッション処理
- [ ] Hand Tracking + Nail Segmentation の初期化
- [ ] 選択デザインデータ（design_data JSON）からDeepAR用ARエフェクトへの変換ロジック
- [ ] リアルタイムARレンダリング（30fps）
- [ ] 爪データ自動検出（長さ・形状・既存ジェル有無）

### 3-2. AR 試着バックエンド ✅
- [x] `POST /ar/sessions` — ARセッション保存（CDNバリデーション付き）
- [x] `GET /ar/sessions/:id` — セッション取得（所有者チェック付き）
- [x] バックエンド: 爪面積・デザイン複雑度 → 施術時間・ジェル量の算出ロジック（`ar_service.go`）
- [x] ARリポジトリ（`ar_repository.go`）

### 3-3. AR 試着 UI（Flutter） ⚠️
- [x] AR試着画面プレースホルダー（Phase 3 着手前の表示）
- [ ] AR試着画面（C-30）— DeepAR カメラビュー・デザイン選択ミニパレット（Phase 3待ち）
- [ ] AR試着結果画面（C-31）— 検出データ表示・SNSシェア・予約導線（Phase 3待ち）

---

## Phase 4: 逆オークション ✅ 完了

### 4-1. 予約リクエスト（Consumer） ✅
- [x] `POST /auctions/requests` — リクエスト作成（3件上限チェック付き）
- [x] `DELETE /auctions/requests/:id` — キャンセル（open/bidding のみ可）
- [x] 予約リクエスト作成画面（C-40）— 予算スライダー・日程ピッカー・エリア選択

### 4-2. サロン入札 ✅
- [x] `GET /auctions/requests` [Salon] — マッチングリクエスト一覧
- [x] `POST /auctions/requests/:id/bids` — 入札（userID→salonID解決済み）
- [x] `PATCH /auctions/requests/:id/bids/:id` — 入札修正（価格引き下げのみ）
- [x] サロンダッシュボード（S-03）— リクエスト一覧カード
- [x] 入札フォーム（S-05）— 価格・日時・割引理由・メッセージ

### 4-3. リアルタイム入札通知 ✅
- [x] SSEハブ実装（Go）— `redis Pub/Sub` チャンネル `auction:{request_id}` でマルチPodブロードキャスト
- [x] `context.Done()` で接続終了時に goroutine クリーンアップ
- [ ] Flutter側 EventSource受信・入札一覧リアルタイム更新（SSE受信は Phase 4 Flutter実装として追加予定）
- [ ] FCM Push通知（アプリ非起動時）

### 4-4. 入札一覧・選択 UI（Flutter） ✅
- [x] 入札一覧画面（C-41）— 有効期限タイマー・キャンセル確認
- [ ] サロン詳細画面（C-42）— ポートフォリオ表示（プレースホルダー）

---

## Phase 5: 決済 & ロイヤリティ分配 ⚠️ 部分完了

### 5-1. Stripe 統合 ⚠️
- [x] `POST /bookings` — BookingService.ConfirmBooking 実装（Payment レコード作成）
- [x] `POST /bookings/:id/complete` — CompleteBooking 実装（ステータス更新）
- [x] `POST /bookings/:id/cancel` — CancelBooking 実装（user/salon 判定付き）
- [ ] Stripe Connect セットアップ（サロンの口座連携）
- [ ] Stripe Payment Intent 実際の呼び出し（現状はプレースホルダーのclient_secret）
- [ ] 施術完了時のキャプチャ実行

### 5-2. ロイヤリティ分配ワーカー ✅
- [x] 施術完了イベント → `royalty_distributions`レコード生成（`royalty_worker.go`）
- [x] 多段フォーク分配計算ロジック（3段打ち切り: depth0=70%, depth1=21%, depth2=9%）
- [x] `UNIQUE(payment_id, user_id)` で SQS at-least-once 冪等性保証
- [x] `design_ip_id NOT NULL` カラム追加済み
- [ ] ユーザーの `point_balance` 更新（ロイヤリティ入金時）
- [ ] サロンへのStripe送金

### 5-3. ウォレット・決済 UI（Flutter） ⚠️
- [x] IPウォレット画面（C-54）— ポイント残高・収益履歴（実データはPhase 5以降）
- [x] 予約確定画面（C-43 相当）— BookingDetail からレビュー + キャンセル
- [ ] Stripe決済UI（Stripe Flutter SDK 統合）
- [ ] 予約完了画面（C-44）

---

## Phase 6: レビュー & バッジシステム ✅ 完了

### 6-1. レビュー バックエンド ✅
- [x] `POST /reviews` — 投稿（CDNバリデーション + DesignIPID解決）
- [x] `GET /salons/:id/reviews` — サロンレビュー一覧
- [ ] SageMaker: before/after画像の類似度スコア算出ワーカー（Phase 6 本実装）
- [ ] サロンの `avg_reproduction_score` リアルタイム更新
- [ ] `skill_badge_tags` の閾値超え時バッジ自動付与ロジック

### 6-2. レビュー UI（Flutter） ✅
- [x] レビュー投稿機能（C-52）— BookingDetailScreen に星評価・コメント投稿シート統合
- [ ] サロンレビュー一覧（サロン詳細画面の一部）

---

## Phase 7: データ基盤（Snowflake） ❌ 未着手

### 7-1. CDC パイプライン構築
- [ ] Debeziumの設定（PostgreSQL WAL読み取り）
- [ ] Kafka Connectセットアップ
- [ ] **Snowflake Kafka Connector**（Confluent製）のセットアップ → Snowflakeへ自動ロード
  - ※ Snowpipeは S3/GCS 等のオブジェクトストレージからの取り込みが主用途。Kafkaからは Snowflake Kafka Connector を使う

### 7-2. Snowflakeスキーマ設計
- [ ] ディメンションテーブル（users / designs / salons）
- [ ] ファクトテーブル（bookings / payments / reviews / royalty）
- [ ] 匿名化ビュー（B2Bデータ外販用）

---

## Phase 8: 最終仕上げ・公開準備 ❌ 未着手

- [ ] 全画面のUI/UXポリッシュ（アニメーション・ローディング・エラーステート）
- [ ] 通知センター画面（C-50〜 画面群）
- [ ] Admin管理画面の整備
- [ ] セキュリティレビュー（SQL injection / 認可漏れ / S3アクセス制御）
- [ ] 負荷テスト（逆オークションSSEの同時接続数）
- [ ] App Store / Google Play の申請準備
- [ ] 多言語対応（i18n）の骨格整備

---

## 実装優先度マトリクス（更新版）

| 機能 | ビジネス価値 | 技術難易度 | 優先度 | 状況 |
|------|------------|-----------|--------|------|
| DB スキーマ + 基本 CRUD | 高 | 低 | P0 | ✅ 完了 |
| 認証基盤 | 高 | 低 | P0 | ✅ 完了 |
| 逆オークション + SSE | 最高（Moat核心） | 高 | P0 | ✅ 完了 |
| ロイヤリティ分配 | 最高（Moat核心） | 中 | P0 | ✅ 完了 |
| レビュー + バッジ | 高（Moat補完） | 中 | P1 | ✅ 完了 |
| AR 試着 + 爪データ抽出 | 最高（Moat核心） | 最高 | P0 | ⚠️ バックエンドのみ |
| 3D エディター | 高 | 高 | P0 | ❌ 未実装 |
| Stripe 決済 | 高 | 中 | P0 | ⚠️ スタブのみ |
| Snowflake CDC | 中（将来Moat） | 高 | P1 | ❌ 未着手 |
| インフラ (EKS/Terraform) | 高 | 高 | P0 | ❌ 未着手 |
| B2B データ外販 | 中 | 高 | P2 | ❌ 未着手 |
| レコメンドAI | 中 | 最高 | P2 | ❌ 未着手 |

---

## 現在未実装で次に着手すべき項目（優先順）

1. **Stripe 本統合**（Phase 5）— `stripe-go` SDK で Payment Intent・キャプチャ・返金を実装
2. **3D エディター**（Phase 2）— Three.js WebView ブリッジから着手
3. **AR / DeepAR**（Phase 3）— Option A（`ar_flutter_plugin`）でPoC開始
4. **AWSインフラ**（Phase 0-5）— Terraform EKS / RDS / ElastiCache
5. **Flutter SSE受信**（Phase 4）— `eventsource` パッケージで入札リアルタイム更新
6. **FCM Push通知**（Phase 4）— `firebase_messaging` でバックグラウンド通知
