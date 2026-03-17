# 開発環境のセットアップ
setup:
    mise install
    pnpm install

# Protobuf のリント
lint:
    buf lint

# コード生成（buf + sqlc）
generate:
    buf generate
    cd gen/go && go mod init github.com/nangashi/bmkr/gen/go 2>/dev/null || true
    cd gen/go && go mod tidy
    cd services/product-mgmt && sqlc generate
    cd services/customer-mgmt && sqlc generate
    cd services/ec-site && sqlc generate

# Docker Compose 起動
db-up:
    docker compose up -d

# Docker Compose 停止
db-down:
    docker compose down

# 全サービスの goose マイグレーション適用
db-migrate:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/product?sslmode=disable" goose -dir services/product-mgmt/db/migrations up
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/customer?sslmode=disable" goose -dir services/customer-mgmt/db/migrations up
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/ecsite?sslmode=disable" goose -dir services/ec-site/db/migrations up

# 全サービスの goose マイグレーション ロールバック
db-rollback:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/product?sslmode=disable" goose -dir services/product-mgmt/db/migrations down
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/customer?sslmode=disable" goose -dir services/customer-mgmt/db/migrations down
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/ecsite?sslmode=disable" goose -dir services/ec-site/db/migrations down

# DB 初期化（volume 削除 → 再起動 → マイグレーション再適用）
db-reset:
    docker compose down -v
    docker compose up -d
    @echo "Waiting for database to be ready..."
    @sleep 5
    just db-migrate

# 全サービス並列起動
dev:
    #!/usr/bin/env bash
    trap 'kill 0' EXIT
    (cd services/product-mgmt && go run main.go) &
    (cd services/customer-mgmt && go run main.go) &
    (cd services/ec-site && go run main.go) &
    (cd services/bff && pnpm dev) &
    wait

# Ory Hydra OAuth 2.0 クライアント登録
hydra-setup:
    bash scripts/hydra-setup.sh
