# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**nailX** は次世代UGC型ネイルプラットフォーム。3つのMoatを中心に設計されている:
1. **デザインIP経済圏** — フォーク&ライセンス構造 + ロイヤリティ自動分配
2. **ユーザー爪質カルテデータ** — AR計測による個人の爪メトリクス蓄積
3. **ネイリスト再現度スコア** — AI解析 + レビュー連動バッジ

3者（デザイナー/消費者/サロン）のエコシステムを逆オークション型マッチングで結ぶ。

---

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| フロントエンド | Flutter 3.x / Dart（iOS 14+ / Android 10+）|
| 状態管理 | Riverpod 2.x |
| HTTP通信 | Dio + SSE（EventSource：入札リアルタイム通知）|
| AR/3D | DeepAR SDK（手トラッキング + 爪セグメンテーション）+ Three.js（WebView経由）|
| バックエンド | Go 1.22+ / Echo v4 |
| 認証 | Firebase Auth（Google / Apple / LINE）+ JWT検証（Go側）|
| DB | PostgreSQL 16 + pgvector（512次元）|
| ORM/クエリ | pgx/v5 + sqlc（型安全クエリ生成）|
| バリデーション | go-playground/validator |
| 非同期処理 | AWS SQS + Worker パターン |
| インフラ | AWS EKS / RDS（Multi-AZ）/ S3 / CloudFront / SageMaker / ECR |
| Analytics DWH | Snowflake（CDC via Debezium → Snowpipe）|
| 決済 | Stripe |
| テスト | testify + gomock |

---

## プロジェクト構造（実装予定）

```
nailX/
├── backend/
│   ├── cmd/server/          # エントリーポイント
│   ├── internal/
│   │   ├── domain/          # ドメインモデル・インターフェース定義
│   │   ├── handler/         # Echoハンドラー（APIルーティング）
│   │   ├── service/         # ビジネスロジック
│   │   ├── repository/      # DB操作（sqlc生成コード利用）
│   │   └── middleware/      # JWT検証・ロギング等
│   ├── migrations/          # golang-migrate SQLファイル
│   └── sqlc.yaml
├── frontend/
│   ├── lib/
│   │   ├── features/        # 機能別ディレクトリ（designs/, auction/, booking/, ar/...）
│   │   ├── core/            # HTTPクライアント・ルーティング・共通ユーティリティ
│   │   └── shared/          # UIコンポーネント・デザインシステム
│   └── pubspec.yaml
├── infra/
│   ├── terraform/
│   └── k8s/
└── claude_context/          # 仕様書群（変更しない）
```

**現状**: ソースコードはまだ存在しない。`claude_context/` と `docs/` に仕様書のみある状態。実装はPhase 0から開始。

---

## よく使うコマンド（実装後）

### バックエンド（Go）
```bash
# ローカルDB起動（PostgreSQL + pgvector）
docker compose up -d db

# DBマイグレーション
migrate -path backend/migrations -database $DATABASE_URL up

# sqlcコード生成
cd backend && sqlc generate

# テスト実行（全体）
cd backend && go test ./...

# 単一パッケージのテスト
cd backend && go test ./internal/service/... -v

# サーバー起動
cd backend && go run cmd/server/main.go
```

### フロントエンド（Flutter）
```bash
# 依存パッケージ取得
cd frontend && flutter pub get

# コード生成（Riverpod / Freezed / json_serializable）
cd frontend && dart run build_runner build --delete-conflicting-outputs

# テスト実行
cd frontend && flutter test

# 単一テストファイル
cd frontend && flutter test test/features/auction/auction_service_test.dart

# iOS シミュレーターで実行
cd frontend && flutter run -d ios

# Android エミュレーターで実行
cd frontend && flutter run -d android
```

---

## アーキテクチャの重要概念

### デザインIPとフォーク・ロイヤリティ
- `design_ips`テーブルに`parent_design_id`（フォーク元）と`similarity_vector`（pgvector 512次元）を持つ
- 類似度検索: コサイン類似度 ≥ 0.90 → 拒否、0.75〜0.90 → フォーク強制、< 0.75 → 独立IP（Admin可変）
- ロイヤリティ分配は**3段打ち切り**: 直接70% / 親21% / 祖父9%（`design_royalty_nodes`で管理）
- IP利用料は施術代の**5%**、プラットフォーム手数料は**15%**

### 逆オークション（予約リクエスト → 入札）
- ユーザーが`booking_requests`を出品 → サロンが`bids`を一発提示（チャットなし）
- リアルタイム通知はSSE（`/auctions/:id/stream`）で配信
- 類似度計算・ロイヤリティ集計はSQS Workerで非同期処理
- 同時オープンリクエスト上限: 3件/ユーザー

### 決済フロー
- 予約確定時にStripe Authorization（オーソリ）
- **施術完了確認後**にキャプチャ → IP利用料分配（`royalty_distributions`生成）
- ポイント残高は初期フェーズではアプリ内クーポンのみ（現金化は法務確認待ち）

### AR爪データ
- DeepAR SDKで21関節 + 爪10枚を検出し`ar_sessions`に保存
- `hand_snapshot_url`（S3写真）は90日で自動削除、メタデータ（爪長さ・形状）は永続保持
- `nail_profiles`に集約したカルテデータが逆オークション時にサロンへ自動添付

---

## 仕様書の場所

全ての設計情報は `claude_context/` に格納されている。実装前に必ず参照すること。

| ファイル | 内容 |
|---------|------|
| `architecture.md` | システム全体構成・サービス分割・データフロー |
| `api_spec.md` | REST APIエンドポイント仕様（70+件）|
| `db_schema.md` | テーブルDDL・インデックス設計 |
| `functional_requirements.md` | 機能要件一覧（Moat貢献度評価付き）|
| `implementation_tasks.md` | 8フェーズの実装ロードマップ（Phase 0〜7）|
| `screen_list.md` | 画面一覧・遷移図 |
| `open_questions.md` | 決定済み事項・保留事項 |

---

## 実装ルール

- **実装はユーザーの明示的な承認後にのみ開始**（`implementation_tasks.md`の原則）
- 各Phaseの完了後にレビューを挟む
- `open_questions.md`の🔴ブロッカー事項はPhase 0完了前に全確定（現時点では全て確定済み）
- sqlcを使用するため、生のクエリ文字列をGoコードに直書きしない
- SSEエンドポイントはgoroutineリークに注意（context cancellationを必ず処理）
