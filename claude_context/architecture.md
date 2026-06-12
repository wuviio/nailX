# システムアーキテクチャ設計書

## サービス概要
次世代UGC型ネイルプラットフォーム。デザインIP経済圏・ユーザー爪質カルテデータ・ネイリスト再現度スコアの3つをMoatとする。

---

## システム全体構成図

```
┌───────────────────────────────────────────────────────────────────────┐
│                    Flutter App (iOS 14+ / Android 10+)                │
│                                                                       │
│  ┌────────────────┐  ┌──────────────────┐  ┌──────────────────────┐   │
│  │  DeepAR SDK    │  │  Three.js (WebGL)│  │   REST API Client    │   │
│  │  Hand Tracking │  │  3D Design Editor│  │   (Dio + SSE)        │   │
│  │  Nail Segment. │  │  Texture Params  │  │                      │   │
│  └────────────────┘  └──────────────────┘  └──────────────────────┘   │
└───────────────────────────────────────────────────────────────────────┘
                                │ HTTPS / REST / SSE
┌───────────────────────────────▼──────────────────────────────────────┐
│                    Go Backend (API Gateway + Services)               │
│                                                                      │
│  ┌──────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────────┐  │
│  │  Auth    │  │ Design IP  │  │  Auction   │  │ Payment & Royalty│  │
│  │ Service  │  │  Service   │  │  Service   │  │     Service      │  │
│  └──────────┘  └────────────┘  └────────────┘  └──────────────────┘  │
│  ┌──────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────────┐  │
│  │ Booking  │  │   Salon    │  │   Notif    │  │   AR / Nail      │  │
│  │ Service  │  │  Service   │  │  Service   │  │  Analysis Svc    │  │
│  └──────────┘  └────────────┘  └────────────┘  └──────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
         │ pgx/v5             │ SQS/events          │ HTTP
┌────────▼─────────┐  ┌───────▼──────────┐  ┌──────▼──────────────────┐
│   PostgreSQL 16  │  │    Snowflake     │  │        AWS              │
│   + pgvector     │  │  Analytics DWH   │  │  ┌──────────────────┐   │
│   Primary DB     │  │  ML Training Data│  │  │ S3 (media store) │   │
│                  │  │  B2B Data Export │  │  │ CloudFront (CDN) │   │
│  CDC (Debezium)  │──▶│  via Snowpipe  │  │  │ SageMaker (AI)   │   │
└──────────────────┘  └──────────────────┘  │  │ EKS (container)  │   │
                                            │  │ SQS (async)      │   │
                                            │  └──────────────────┘   │
                                            └─────────────────────────┘
```

---

## レイヤー別技術詳細

### 1. フロントエンド（Flutter / Dart）

| 項目 | 内容 |
|------|------|
| フレームワーク | Flutter 3.x (Dart) |
| 対象OS | iOS 14+ / Android 10+ |
| AR/3D | DeepAR SDK（Hand Tracking + Nail Segmentation）|
| 3Dエディター | Three.js（Flutter WebView経由 or flutter_gl） |
| 状態管理 | Riverpod 2.x |
| HTTP通信 | Dio（REST）+ EventSource（SSE: 入札リアルタイム通知）|
| ローカルDB | Hive（オフラインキャッシュ）|
| プッシュ通知 | Firebase Cloud Messaging (FCM) |
| 認証 | Firebase Auth（Google / Apple / LINE ログイン対応）|

**3Dエディターの動作方式:**
- Flutter WebView内でThree.jsを動かし、パーツ配置データをJSON形式でFlutter側と双方向通信
- デザイン確定時にパーツ配置JSON（`design_data`）をバックエンドへ送信

**AR試着の動作方式:**
- DeepAR SDKでカメラ映像をリアルタイム処理
- 手の21関節 + 爪10枚のセグメンテーションを検出
- 選択デザインの3DオブジェクトをAR合成してレンダリング
- 撮影完了時に爪データ（長さ・形状・既存ジェル有無）をバックエンドへ送信

---

### 2. バックエンド（Go）

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.22+ |
| HTTPフレームワーク | Echo v4 |
| 認証 | JWT（Firebase Auth ID Token検証）|
| DBドライバー | pgx/v5 + sqlc（型安全クエリ生成）|
| バリデーション | go-playground/validator |
| リアルタイム | SSE（Server-Sent Events）：入札通知。EKS複数Pod間は **Redis Pub/Sub（ElastiCache）** でブロードキャスト |
| 非同期処理 | AWS SQS + Workerパターン（類似度計算・ロイヤリティ集計）|
| テスト | testify + gomock |

**サービス分割（論理）:**

```
api/
├── auth/          # JWT検証ミドルウェア、Firebase連携
├── users/         # ユーザーCRUD、爪カルテ管理
├── designs/       # デザインIP登録・フォーク・フィード
├── ar/            # AR試着セッション保存・爪データ解析
├── salons/        # サロンCRUD、ポートフォリオ
├── auctions/      # 予約リクエスト・入札・SSEハブ
├── bookings/      # 予約確定・キャンセル
├── payments/      # Stripe連携・ロイヤリティ分配
├── reviews/       # レビュー投稿・バッジ更新
└── notifications/ # プッシュ通知・通知履歴
```

---

### 3. データベース（PostgreSQL 16 + pgvector）

| 項目 | 内容 |
|------|------|
| バージョン | PostgreSQL 16 |
| 拡張 | pgvector（512次元ベクトル：デザイン類似度検索）|
| マイグレーション | golang-migrate |
| 接続プール | pgxpool（最大接続数: 環境変数で制御）|

**pgvectorの用途:**
- デザインIP登録時にSageMakerでembedding（512次元）を生成
- `<=>` 演算子（コサイン距離）で類似デザインを検索。インデックスは `HNSW + vector_cosine_ops`
- コサイン類似度スコア（0〜1）で閾値判定 → 「盗作」として弾く / 「派生（フォーク）」として登録許可

---

### 4. データ分析基盤（Snowflake）

| 項目 | 内容 |
|------|------|
| 連携方式 | CDC: PostgreSQL → Debezium → Kafka → **Snowflake Kafka Connector** → Snowflake |
| 用途① | マッチングAI学習データ（ユーザー × デザイン × サロン × 結果）|
| 用途② | ネイルメーカー向けB2Bデータ外販（匿名化済み）|
| 用途③ | エリア別・性別別トレンド分析 |

---

### 5. インフラ（AWS）

| リソース | 用途 |
|----------|------|
| EKS | Goバックエンドコンテナのオーケストレーション |
| RDS (PostgreSQL) | 本番DB（Multi-AZ）|
| S3 | 3D素材・AR試着写真・施術before/after画像 |
| CloudFront | S3メディアのCDN配信 |
| SageMaker | デザイン類似度embeddingモデルのホスティング |
| SQS | 非同期処理キュー（類似度計算・ロイヤリティ分配）|
| ElastiCache (Redis) | SSE マルチPodブロードキャスト用 Pub/Sub |
| ECR | Dockerイメージレジストリ |
| Secrets Manager | DB接続情報・APIキー管理 |

---

## 主要データフロー

### フロー①: デザイン作成 → IP登録

```
[Flutter 3Dエディター]
  └─ パーツ配置JSON送信
       ↓
[Go: /api/designs (POST)]
  ├─ design_dataをDBに仮保存 (status: 'pending')
  ├─ SQSへ類似度計算ジョブをエンキュー
  └─ ジョブID返却
       ↓
[Go Worker: 類似度計算]
  ├─ SageMaker呼び出し → embedding(512次元) 生成
  ├─ PostgreSQL(pgvector): 類似デザイン検索
  │     ├─ 類似度スコア < 閾値A → 独立IPとして承認 (status: 'active')
  │     ├─ 閾値A ≤ スコア < 閾値B → フォークIPとして承認
  │     │      └─ design_royalty_nodes に分配ツリー生成
  │     └─ スコア ≥ 閾値B → 盗作として拒否 (status: 'rejected')
  └─ SSE/Push通知でユーザーへ結果通知
       ↓
[Snowflake CDC連携]
  └─ design_ipsテーブル変更をSnowpipeで取り込み
```

### フロー②: 逆オークション（AR試着〜予約確定）

```
[Flutter AR試着]
  ├─ DeepAR: 爪長さ・形状・既存ジェル有無を自動検出
  └─ AR session + 爪データ → Go: /api/ar/sessions (POST)
       ↓
[Flutter 予約リクエスト作成]
  └─ design_ip_id / 予算 / 希望日時 / エリア → Go: /api/auctions/requests (POST)
       ↓
[Go: 対象サロンへSSE/Push通知]
  └─ エリア・スキルタグ・爪データでフィルタリングされたサロンへ通知
       ↓
[サロン: 入札]
  └─ price / slot / includes_removal → Go: /api/auctions/bids (POST)
       ↓
[ユーザー: 入札一覧閲覧 → 選択]
  └─ bid_id → Go: /api/bookings (POST) → booking確定
       ↓
[Stripe決済]
  └─ 決済完了 → payments, royalty_distributions レコード生成
              → サロン送金 / IP利用料分配（非同期）
```

### フロー③: レビュー → バッジ更新

```
[施術完了後: Flutter]
  └─ 再現度スコア + after画像 → Go: /api/reviews (POST)
       ↓
[Go Worker]
  ├─ SageMaker: before/after画像の類似度スコア算出（AI再現度）
  ├─ サロンのavg_reproduction_scoreを再計算
  └─ skill_badge_tagsを更新（閾値超え時にバッジ付与）
       ↓
[Snowflake]
  └─ ネイリストスキルデータをDWHに連携
```

---

## SSE マルチPod ブロードキャスト設計

EKS 上で複数 Pod が動作する場合、入札書き込み Pod と SSE 接続 Pod が異なることがある。  
同一 Pod 内にしかブロードキャストしないと、別 Pod のクライアントへ通知が届かない。

```
[入札 POST /bids]
  └─ Go Pod A: DB に bid 書き込み
       ↓
     Redis Publish → channel: "auction:{request_id}"
       ↓
  ┌───────────────┬───────────────┐
Go Pod A        Go Pod B        Go Pod C
(SSE Subscribe) (SSE Subscribe) (SSE Subscribe)
  ↓               ↓               ↓
接続クライアントへ転送（各Podが担当する接続のみ）
```

- **実装**: 各 Pod 起動時に当該チャンネルを `SUBSCRIBE`。入札書き込み側は `PUBLISH` のみ
- **ElastiCache**: Redis クラスター（Cluster Mode Disabled、Single AZ でもPub/Subは機能する）
- **goroutine リーク対策**: SSE ハンドラーは `context.Done()` を確認して Redis Subscriber と goroutine を確実にクリーンアップ

---

## セキュリティ設計

| 項目 | 方針 |
|------|------|
| 認証 | Firebase Auth ID Token（Bearerトークン）|
| 認可 | ロールベース（consumer / salon_owner / admin）|
| AR画像 | S3保存・署名付きURL（有効期限付き）・暗号化 |
| 決済 | カード情報はStripeが保持（PCI DSS準拠）|
| 個人情報 | 爪データは匿名化してSnowflakeへ連携 |
| APIレート制限 | IP単位 + ユーザー単位（Echoミドルウェア）|
