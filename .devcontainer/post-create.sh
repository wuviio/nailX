#!/usr/bin/env bash
# Dev Container 初回作成時に 1 回だけ実行されるセットアップスクリプト
set -euo pipefail

WORKSPACE=/workspace
echo "========================================="
echo " nailX Dev Container post-create setup"
echo "========================================="

# -------- backend/.env --------
if [ ! -f "${WORKSPACE}/backend/.env" ]; then
  cp "${WORKSPACE}/backend/.env.example" "${WORKSPACE}/backend/.env"
  echo "[OK] backend/.env を .env.example からコピーしました"
  echo "     FIREBASE_CREDENTIALS_PATH / STRIPE_SECRET_KEY などを手動で設定してください"
else
  echo "[SKIP] backend/.env はすでに存在します"
fi

# -------- Go ツールインストール --------
echo ""
echo "--- Go ツールをインストールしています ---"
# gopls: Language Server
go install golang.org/x/tools/gopls@latest
# dlv: デバッガー
go install github.com/go-delve/delve/cmd/dlv@latest
# sqlc: 型安全クエリ生成（Phase 2 以降で利用）
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.26.0
# air: ホットリロード（任意）
go install github.com/cosmtrek/air@latest
echo "[OK] Go ツールのインストール完了"

# -------- backend Go 依存解決 --------
echo ""
echo "--- backend: go mod download & tidy ---"
cd "${WORKSPACE}/backend"

# Go モジュールキャッシュディレクトリを作成（既に devcontainer.json で GOMODCACHE が設定されている）
# ただし、ボリュームマウント直後の初期化の遅延に備えて明示的に作成
mkdir -p "${GOMODCACHE:=${HOME}/.cache/go-mod}"

if go mod download; then
  echo "[OK] Go モジュールダウンロード完了"
else
  echo "[WARN] go mod download に失敗。ネットワーク接続を確認し、後で手動で以下を実行してください："
  echo "       GOMODCACHE=${GOMODCACHE} GOSUMDB=off go mod tidy"
fi

# go.sum が存在しない or 不完全な場合に備えて tidy も実行
if go mod tidy; then
  echo "[OK] go mod tidy 完了。go.sum が生成されました"
else
  echo "[WARN] go mod tidy に失敗。.env に秘匿値が設定されていない可能性があります。"
  echo "       .env を編集した後、以下を手動で実行してください："
  echo "       GOMODCACHE=${GOMODCACHE} GOSUMDB=off go mod tidy"
fi

# -------- frontend pub get --------
echo ""
echo "--- frontend: flutter pub get ---"
cd "${WORKSPACE}/frontend"

# pub-cache をワークスペース内に配置（/home/vscode/.pub-cache は root 所有で書き込み不可）
export PUB_CACHE="${WORKSPACE}/nailX/.pub-cache"
mkdir -p "${PUB_CACHE}"

if flutter pub get; then
  echo "[OK] Flutter パッケージ取得完了"
else
  echo "[WARN] flutter pub get に失敗。ネットワーク接続を確認し、後で手動で以下を実行してください："
  echo "       cd ${WORKSPACE}/frontend && flutter pub get"
fi

# -------- flutter コード生成（Freezed / Riverpod / json_serializable）--------
echo ""
echo "--- frontend: dart run build_runner build ---"
cd "${WORKSPACE}/frontend"

if dart run build_runner build --delete-conflicting-outputs; then
  echo "[OK] Dart コード生成完了（.g.dart / .freezed.dart）"
else
  echo "[WARN] dart run build_runner build に失敗。以下の理由が考えられます："
  echo "  1. .env に秘匿値（FIREBASE_CREDENTIALS_PATH など）が設定されていない"
  echo "  2. Go モジュール依存解決が未完了（pubspec.lock が古い）"
  echo ""
  echo "対処方法："
  echo "  1. FIREBASE_CREDENTIALS_PATH / STRIPE_SECRET_KEY などを .env に設定"
  echo "  2. 以下を手動で実行："
  echo "     cd ${WORKSPACE}/frontend && flutter pub get"
  echo "     cd ${WORKSPACE}/frontend && dart run build_runner build --delete-conflicting-outputs"
fi

# -------- flutter doctor（診断情報） --------
echo ""
echo "--- flutter doctor (Web のみ表示) ---"
flutter doctor --verbose 2>&1 | grep -A 3 "Chrome\|Web" || true

echo ""
echo "========================================="
echo " セットアップ完了！"
echo "========================================="
echo ""
echo "📌 上記で [WARN] が出た場合："
echo "   1. backend/.env に秘匿値を設定する"
echo "   2. 以下を手動で実行する："
echo ""
echo "     # backend"
echo "     cd ${WORKSPACE}/backend"
echo "     GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go mod tidy"
echo ""
echo "     # frontend"
echo "     cd ${WORKSPACE}/frontend"
echo "     flutter pub get"
echo "     dart run build_runner build --delete-conflicting-outputs"
echo ""
echo "📖 詳細は docs/startup-guide.md を参照してください"
echo "========================================="
