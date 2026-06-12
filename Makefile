.PHONY: up down migrate-up migrate-down server lint test

# ローカル DB + Redis 起動
up:
	docker-compose up -d postgres redis

# 起動 + マイグレーション実行
up-full:
	docker-compose up -d

# コンテナ停止
down:
	docker-compose down

# マイグレーション（golang-migrate が必要: brew install golang-migrate）
migrate-up:
	migrate -path backend/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path backend/migrations -database "$$DATABASE_URL" down 1

# バックエンド起動
server:
	cd backend && GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go run ./cmd/server

# lint（golangci-lint が必要）
lint:
	cd backend && GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off golangci-lint run ./...

# テスト
test:
	cd backend && GOMODCACHE=/workspace/.go/pkg/mod GOSUMDB=off go test ./...

# flutter コード生成
flutter-gen:
	cd frontend && flutter pub run build_runner build --delete-conflicting-outputs
