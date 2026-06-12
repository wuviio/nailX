# nailX 起動手順ガイド

**対象環境:** Windows 11 + VS Code + Dev Container  
**最終更新:** 2026-06-11

---

## 目次

1. [前提条件と初回インストール](#1-前提条件と初回インストール)
2. [Dev Container の初回セットアップ](#2-dev-container-の初回セットアップ)
3. [毎日の起動手順（通常フロー）](#3-毎日の起動手順通常フロー)
4. [バックエンド起動手順](#4-バックエンド起動手順)
5. [フロントエンド起動手順](#5-フロントエンド起動手順)
6. [秘匿値・環境変数の設定](#6-秘匿値環境変数の設定)
7. [よく使うコマンド集](#7-よく使うコマンド集)
8. [トラブルシューティング](#8-トラブルシューティング)

---

## 1. 前提条件と初回インストール

### 1-1. ホストマシン（Windows 11）に必要なもの

| ツール | バージョン目安 | インストール先 |
|--------|--------------|-------------|
| Docker Desktop | 4.x 以上 | https://www.docker.com/products/docker-desktop/ |
| VS Code | 最新安定版 | https://code.visualstudio.com/ |
| Dev Containers 拡張機能 | 最新 | VS Code 拡張機能マーケットプレイス |
| Flutter SDK（モバイル実機デバッグ用） | 3.19 以上 | https://docs.flutter.dev/get-started/install/windows |
| Git | 2.40 以上 | https://git-scm.com/ |

> **注意:** Flutter SDK のホスト側インストールは iOS/Android 実機・エミュレーター向けデバッグにのみ必要。  
> コード編集・テスト・Web 実行はコンテナ内 Flutter で完結する。

### 1-2. Docker Desktop の設定確認

1. Docker Desktop を起動し、右下のアイコンが緑であることを確認
2. **Settings → General** で「Use WSL 2 based engine」が ON であることを確認（Windows の場合）
3. **Settings → Resources → WSL Integration** で使用中のディストリビューションが有効になっていることを確認
4. **Settings → Resources → Memory** を **8 GB 以上** に設定（不足するとビルドが失敗する）

### 1-3. Dev Containers 拡張機能のインストール

VS Code で `Ctrl+Shift+X` を押し、`Dev Containers` を検索してインストール。  
ID: `ms-vscode-remote.remote-containers`

---

## 2. Dev Container の初回セットアップ

> **所要時間:** 初回は Docker イメージビルド + Flutter SDK クローン + Go モジュール DL + Dart コード生成のため **15〜25 分** かかる。2 回目以降はキャッシュが効き 1〜2 分で起動する。

### Step 1: リポジトリをクローン

```bash
# ホストの PowerShell / コマンドプロンプトで実行
git clone <repo-url> D:\PROJECT2026\nailX\nailX
cd D:\PROJECT2026\nailX\nailX
```

### Step 2: VS Code でフォルダを開く

```bash
code .
```

または VS Code の「ファイル → フォルダを開く」で `D:\PROJECT2026\nailX\nailX` を選択。

### Step 3: Dev Container で再度開く

VS Code の右下にポップアップが表示される場合は「**Reopen in Container**」をクリック。  
表示されない場合:

1. `Ctrl+Shift+P` でコマンドパレットを開く
2. `Dev Containers: Reopen in Container` を検索して実行

### Step 4: ビルドの進行を確認

- 左下が `><` から `Dev Container: nailX` に変わったら接続完了
- 初回は ターミナルに `post-create.sh` の出力が流れる（セットアップ自動実行）
- セットアップ完了内容:
  - Go ツール（gopls / dlv / sqlc / air）のインストール
  - `go mod download && go mod tidy`（Go 依存解決・go.sum 生成）
    - コンテナ環境の権限制約により、以下の環境変数が自動設定されます：
      ```bash
      GOMODCACHE=/workspace/.go/pkg/mod    # プロジェクト内のローカルキャッシュを使用
      GOSUMDB=off                          # チェックサム検証を無効化（オフライン開発対応）
      ```
    - Makefile にもこれらが設定されているため、`make server` / `make lint` / `make test` を実行する際は自動適用されます
  - `flutter pub get`（Flutter パッケージ取得）
  - `dart run build_runner build`（.freezed.dart / .g.dart の自動生成）
- 最後に `セットアップ完了！` と表示されればOK

> **⚠️ post-create.sh が失敗・部分実行する理由：**
>
> `post-create.sh` は初回起動時に自動実行されますが、以下の理由で完全には成功しない場合があります：
>
> 1. **秘匿値がまだ設定されていない**
>    - Firebase / AWS / Stripe の認証情報は Step 5 で設定するため、初回の `go mod tidy` や `dart run build_runner` は秘匿値なしで実行される
>    - 後で `.env` を編集した後、コマンドを手動で再実行することで、秘匿値を参照する処理が正常に完了する
>
> 2. **ネットワーク接続の一時的な遅延**
>    - Flutter SDK クローン、Go モジュールダウンロード、pub.dev への接続がタイムアウトする可能性がある
>    - 再実行時はキャッシュが部分的に利用され、残りのダウンロードだけで済む
>
> 3. **Docker のリソース枯渇**
>    - 初回はイメージビルド + 複数ツールのインストール + パッケージダウンロードが同時に動作
>    - メモリ不足でいくつかのタスクが失敗し、その後のステップがスキップされる可能性
>
> **対処方法：** Step 5 で `.env` を編集した後、以下のコマンドを手動で再実行してください
> ```bash
> cd /workspace/backend
> GOMODCACHE=~/.cache/go-mod GOSUMDB=off go mod tidy
> 
> cd /workspace/frontend
> flutter pub get
> dart run build_runner build --delete-conflicting-outputs
> ```

### Step 5: .env を確認・編集

```bash
# コンテナ内のターミナル（VS Code のターミナルタブ）で実行
cat /workspace/backend/.env
```

`FIREBASE_CREDENTIALS_PATH` / `STRIPE_SECRET_KEY` などの秘匿値を設定する（後述の[第6節](#6-秘匿値環境変数の設定)を参照）。

> **重要：** `.env` を編集した後、以下を手動で実行してください。これにより、秘匿値を参照する処理（Go ビルド、Dart コード生成など）が正常に完了します。
>
> ```bash
> cd /workspace/backend
> GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go mod tidy
> 
> cd /workspace/frontend
> flutter pub get
> dart run build_runner build --delete-conflicting-outputs
> ```
>
> 理由は上記「Step 4」の注釈を参照してください。---

## 3. 毎日の起動手順（通常フロー）

2 回目以降の起動はキャッシュが効くため短時間で完了する。

```
[ホスト] Docker Desktop が起動していることを確認
    ↓
[VS Code] フォルダを開く → Dev Container で開き直す（自動で接続）
    ↓
[コンテナ内] DB・Redis が起動しているか確認
    ↓
[コンテナ内] バックエンドサーバーを起動
    ↓
[ホスト or コンテナ内] フロントエンドを起動
```

### 確認コマンド（コンテナ内ターミナル）

```bash
# PostgreSQL が応答するか確認（postgresql-client はコンテナ内にインストール済み）
psql "$DATABASE_URL" -c "SELECT 1"

# Redis が応答するか確認（redis-tools はコンテナ内にインストール済み）
redis-cli -u "$REDIS_URL" PING   # → PONG が返ればOK
```

> `DATABASE_URL` / `REDIS_URL` は Dev Container 起動時に自動で設定される（`postgres` / `redis` のサービス名でアクセス）。

---

## 4. バックエンド起動手順

すべてコンテナ内ターミナルで実行する（`/workspace/backend` がワーキングディレクトリ）。

### Step 1: ディレクトリ移動

```bash
cd /workspace/backend
```

### Step 2: DB マイグレーション実行

```bash
# スキーマを最新状態に適用する
migrate -path ./migrations -database "$DATABASE_URL" up
```

> **説明:**
> - `migrations/000001_init.up.sql` が適用され、全テーブル・インデックス・トリガーが作成される
> - `docker-compose` 起動時に `migrate` サービスが自動実行されるため、すでに適用済みの場合は「no change」と表示される（冪等）
> - `migrate up N` で N ステップだけ進めることも可能

エラーが出た場合:

```bash
# マイグレーションのバージョン状態を確認
migrate -path ./migrations -database "$DATABASE_URL" version

# dirty 状態をリセットしてからやり直す（開発環境のみ）
migrate -path ./migrations -database "$DATABASE_URL" force 1
migrate -path ./migrations -database "$DATABASE_URL" up
```

### Step 3: サーバー起動

```bash
# 環境変数を .env から読み込んでサーバーを起動
# DATABASE_URL / REDIS_URL はコンテナ起動時から設定済みだが .env でも同値を指定している
set -a && source .env && set +a
go run ./cmd/server
```

または Makefile を使った方法（環境変数が自動設定される）:

```bash
# cd /workspace/backend で実行
make server
```

> **環境変数の説明:**
> - `GOMODCACHE=/workspace/.go/pkg/mod` ：Go モジュールキャッシュをプロジェクト内に配置（コンテナ内権限制約対応）
> - `GOSUMDB=off` ：チェックサム検証を無効化（オフライン・ローカル開発対応）
> - これらは `devcontainer.json` の `containerEnv` で自動設定されるため、コンテナ起動後のすべてのプロセスで有効です
> - Makefile にも同じ値が設定されているため、通常は意識不要

> **説明:**
> - `.env` の `DATABASE_URL` はコンテナ内のサービス名 `postgres` を使うよう設定されている（`localhost` ではない）
> - ポート 8080 で Listen が始まると `{"level":"INFO","msg":"server started","port":"8080"}` が表示される
> - VS Code が自動でポート 8080 をホストへフォワードし、ブラウザから `http://localhost:8080/health` でアクセスできる

**起動確認:**

```bash
# 別のターミナルタブで実行（Ctrl+Shift+` で新規タブ）
curl http://localhost:8080/health
# → {"status":"ok"}
```

### Step 4: ホットリロード（任意）

`go run` は変更のたびに手動で再起動が必要。`air`（コンテナ内にインストール済み）を使うと自動リロードできる:

```bash
# .env を読み込んでから air を起動
cd /workspace/backend
set -a && source .env && set +a
air
```

または Makefile を使った方法:

```bash
# cd /workspace/backend で実行
make air  # (Makefile に air タスクがあれば)
```

> `air` は code.hot_reload_delay_time_ms などの設定を `.air.toml` で管理できます（ビルド遅延の最適化に便利）

### Step 5: テスト実行

```bash
# 全テスト
go test ./...

# 特定パッケージだけ（詳細出力）
go test ./internal/service/... -v -run TestAuction

# カバレッジ計測
go test ./... -coverprofile=cover.out && go tool cover -html=cover.out
```

### Step 6: Lint

```bash
golangci-lint run ./...
```

---

## 5. フロントエンド起動手順

### 5-A. Flutter Web（コンテナ内で完結）

コンテナ内に Flutter SDK がインストールされているため、Web ビルドはコンテナ内で動作する。

```bash
cd /workspace/frontend
```

**依存パッケージの確認:**

```bash
flutter pub get
```

**コード生成（Riverpod / Freezed / json_serializable）:**

```bash
dart run build_runner build --delete-conflicting-outputs
```

> **説明:**
> - `post-create.sh` で初回自動実行されるが、モデルクラスを変更したときは再実行が必要
> - 生成対象: `@riverpod` プロバイダー → `.g.dart`、`@freezed` データクラス → `.freezed.dart` + `.g.dart`

**Flutter Web で実行:**

```bash
flutter run -d web-server --web-port 3000 --web-hostname 0.0.0.0
```

> **説明:**
> - `web-server` デバイスは Chrome なしで HTTP サーバーとして起動する（ヘッドレス）
> - `--web-hostname 0.0.0.0` でコンテナ外（ホスト）からアクセスできるようにする
> - VS Code がポート 3000 を自動フォワードし、ブラウザで開く通知が表示される
> - ホストブラウザで `http://localhost:3000` を開くとアプリが表示される
> - コード変更後に `r` キーを押すとホットリロード、`R` でフルリスタート

> **ポートフォワードが通知されない場合:**  
> VS Code の「ポート」タブ（`Ctrl+Shift+P` → `View: Toggle Ports`）でポート 3000 が転送されているか確認。  
> されていない場合は「ポートの転送」→ ポート番号 3000 を手動追加。

### 5-B. Flutter iOS / Android（ホストマシンで実行）

iOS/Android の実機・エミュレーターはホスト側の Flutter SDK が必要。**コンテナ内では実行できない。**

**ホストの PowerShell / ターミナルで実行:**

```powershell
# リポジトリの frontend ディレクトリへ移動
cd D:\PROJECT2026\nailX\nailX\frontend

# 依存解決
flutter pub get

# コード生成（.g.dart / .freezed.dart が未生成の場合）
dart run build_runner build --delete-conflicting-outputs

# Android エミュレーターで実行
flutter run -d android --dart-define=API_BASE_URL=http://10.0.2.2:8080

# iOS シミュレーターで実行（macOS のみ）
flutter run -d ios --dart-define=API_BASE_URL=http://localhost:8080

# 接続中のデバイス一覧を確認
flutter devices
```

> **注意:**
> - バックエンド API はコンテナ内の 8080 ポートがホストへフォワードされるため、ホストからは `http://localhost:8080` でアクセスできる
> - Android エミュレーターからの `localhost` は `10.0.2.2` でアクセスする（AVD の仕様）

### 5-C. Flutter テスト（コンテナ内）

```bash
cd /workspace/frontend

# 全テスト
flutter test

# 単一ファイル
flutter test test/features/auction/auction_service_test.dart

# 静的解析
flutter analyze
```

---

## 6. 秘匿値・環境変数の設定

`backend/.env` は `.gitignore` 対象（コミットしない）。初回セットアップで `.env.example` からコピーされる。

### 設定が必要な項目

| 変数名 | 取得方法 | 説明 |
|--------|---------|------|
| `FIREBASE_CREDENTIALS_PATH` | Firebase Console → プロジェクト設定 → サービスアカウント → 鍵を生成 | Firebase Admin SDK のサービスアカウントJSONファイルのパス |
| `AWS_ACCESS_KEY_ID` | AWS IAM コンソール | S3 / SQS アクセス用 |
| `AWS_SECRET_ACCESS_KEY` | 同上 | 上記のシークレット |
| `STRIPE_SECRET_KEY` | Stripe ダッシュボード → 開発者 → APIキー | テスト環境は `sk_test_` から始まる |
| `STRIPE_WEBHOOK_SECRET` | Stripe ダッシュボード → Webhook | ローカルは `stripe listen` CLI で取得 |

### 主要な環境変数と main.go の対応表

| `.env` 変数名 | `main.go` で読む方法 | 備考 |
|---|---|---|
| `DATABASE_URL` | `mustEnv("DATABASE_URL")` | コンテナ内は `@postgres:5432` を使用 |
| `REDIS_URL` | `mustEnv("REDIS_URL")` | コンテナ内は `redis://redis:6379/0` を使用 |
| `S3_BUCKET` | `envOr("S3_BUCKET", "nailx-media-dev")` | AWS S3 バケット名 |
| `SQS_SIMILARITY_QUEUE_URL` | `os.Getenv("SQS_SIMILARITY_QUEUE_URL")` | 空欄でも起動可（dev モード）|
| `CDN_BASE_URL` | `os.Getenv("CDN_BASE_URL")` | presigned URL のベース |
| `PORT` | `envOr("PORT", "8080")` | 省略時 8080 |

### Firebase 認証ファイルの配置

```bash
# コンテナ内で実行
# firebase-credentials.json を backend/ に配置する
# ※ このファイルは絶対にコミットしない
ls /workspace/backend/firebase-credentials.json   # 存在確認

# .env の設定（すでに .env.example から設定済みのはず）
grep FIREBASE /workspace/backend/.env
```

---

## 7. よく使うコマンド集

コンテナ内ターミナル `/workspace` で実行する。

### バックエンド

```bash
# サーバー起動（.env 読み込み込み）
cd backend && set -a && source .env && set +a && go run ./cmd/server

# または Makefile で環境変数が自動設定される
make server

# ホットリロード起動
cd backend && set -a && source .env && set +a && air

# マイグレーション：進む
migrate -path backend/migrations -database "$DATABASE_URL" up

# マイグレーション：1つ戻す
migrate -path backend/migrations -database "$DATABASE_URL" down 1

# マイグレーション：バージョン確認
migrate -path backend/migrations -database "$DATABASE_URL" version

# テスト（環境変数が自動設定される）
make test
# または手動で環境変数を設定
cd backend && GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go test ./...

# lint（環境変数が自動設定される）
make lint
# または手動で環境変数を設定
cd backend && GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off golangci-lint run ./...

# ビルド確認（コンパイルエラーチェック）
cd backend && go build ./...

# psql で直接クエリ（postgresql-client はインストール済み）
psql "$DATABASE_URL"

# Redis 接続確認（redis-tools はインストール済み）
redis-cli -u "$REDIS_URL" PING
```

### フロントエンド

```bash
# パッケージ取得
cd frontend && flutter pub get

# コード生成（変更後に必ず実行）
cd frontend && dart run build_runner build --delete-conflicting-outputs

# watch モード（ファイル変更で自動再生成）
cd frontend && dart run build_runner watch --delete-conflicting-outputs

# テスト
cd frontend && flutter test

# 解析
cd frontend && flutter analyze

# Web で起動
cd frontend && flutter run -d web-server --web-port 3000 --web-hostname 0.0.0.0
```

### DB 操作

```bash
# psql 接続
psql "$DATABASE_URL"

# テーブル一覧
psql "$DATABASE_URL" -c "\dt"

# 全データ削除してスキーマ再適用（開発リセット）
migrate -path backend/migrations -database "$DATABASE_URL" down --all
migrate -path backend/migrations -database "$DATABASE_URL" up
```

---

## 8. トラブルシューティング

### Q: Dev Container がビルドエラーで起動しない

```
Error response from daemon: ...
```

**原因と対処:**

1. Docker Desktop が起動していない → タスクトレイから起動
2. Docker のメモリ不足 → Settings → Resources でメモリを 8GB 以上に増やす
3. Flutter clone がタイムアウト → ネットワーク接続を確認し `Ctrl+Shift+P` → `Dev Containers: Rebuild Container`

### Q: `migrate up` で `error: dial tcp: connection refused`

**原因:** postgres コンテナがまだヘルスチェックを通過していない。

```bash
# postgres の状態確認
docker ps | grep postgres
# STATUS が "healthy" になるまで待つ（30 秒程度）
```

### Q: `go run ./cmd/server` でビルドエラーが出る

**確認手順:**

```bash
cd /workspace/backend
go build ./...
```

エラーが出た場合は `go mod tidy` を実行:

```bash
# 環境変数を設定してから実行（重要）
GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go mod tidy
```

> **注意:** デフォルトの `/go/pkg/mod/cache` ディレクトリは Dev Container 内で権限がないため、`/workspace/.go/pkg/mod` を使う必要があります。  
> これは devcontainer.json の `containerEnv` に設定されているため、通常は自動で適用されます。

### Q: `mkdir /go/pkg/mod/cache: permission denied` エラーが出る

**原因:** Go のデフォルトキャッシュディレクトリ `/go/pkg/mod/cache` がコンテナ内で書き込み不可。

**対処:**

```bash
# devcontainer.json で自動設定されるため、通常は出現しません
# ただし、環境変数を明示的に上書きしている場合は以下を実行：
export GOMODCACHE=/workspace/.go/pkg/mod
export GOSUMDB=off

# その後で go コマンドを実行
go mod tidy
go run ./cmd/server
```

> **自動対応:** `devcontainer.json` の `containerEnv` に `GOMODCACHE=/workspace/.go/pkg/mod` が設定されているため、通常のコンテナ起動では自動で有効です。  
> Makefile (`make server` / `make test` / `make lint`) を使うと確実です。

### Q: サーバーが `database ping failed` で起動しない

**原因:** `DATABASE_URL` が `localhost` を指している（コンテナ内では `postgres` を使う必要がある）。

**確認:**

```bash
echo $DATABASE_URL
# → postgres://nailx:nailx_dev@postgres:5432/nailx?sslmode=disable  ← 正しい
# → postgres://nailx:nailx_dev@localhost:5432/nailx?sslmode=disable ← コンテナ内では誤り
```

`.env` を確認し、`DATABASE_URL` のホスト名が `postgres`（Docker Compose サービス名）になっていることを確認する。  
古い `.env` を使っている場合は `.env.example` から再コピーしてください。

### Q: Flutter Web で `localhost:3000` につながらない

**確認手順:**

1. VS Code の「ポート」タブ（`Ctrl+Shift+P` → `View: Toggle Ports`）でポート 3000 が転送されているか確認
2. されていない場合は「ポートの転送」→ ポート番号 3000 を手動追加
3. コンテナ内で `flutter run -d web-server --web-port 3000 --web-hostname 0.0.0.0` を実行していることを確認（`--web-hostname 0.0.0.0` が必要）

### Q: `dart run build_runner` で `Conflict: ...` エラー

```bash
# 生成ファイルをすべて削除してから再生成
flutter pub run build_runner clean
flutter pub run build_runner build --delete-conflicting-outputs
```

### Q: `go test` で Firebase 初期化エラーが出る

**原因:** テストコードが Firebase を初期化しようとしているが、サービスアカウント JSON がない。

**対処:** サービスのユニットテストは gomock でインターフェースをモックすることで Firebase 不要になる。

### Q: コンテナを作り直したい（クリーンビルド）

```
Ctrl+Shift+P → Dev Containers: Rebuild Container Without Cache
```

> これを実行すると Docker イメージを最初からビルドし直す。所要時間: 15〜25 分。
