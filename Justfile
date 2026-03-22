# 開発環境のセットアップ
setup:
    mise install
    pnpm install
    lefthook install

# --- Lint ---
# 全 lint 実行
lint: lint-proto lint-ts lint-go

# Protobuf lint
lint-proto:
    buf lint

# TypeScript lint (oxlint)
lint-ts:
    oxlint --config oxlint.json services/bff/src/
    oxlint --config oxlint.json services/ec-site/frontend/src/

# Go lint (golangci-lint)
lint-go:
    cd services/product-mgmt && golangci-lint run ./...
    cd services/customer-mgmt && golangci-lint run ./...
    cd services/ec-site && golangci-lint run ./...

# --- Test ---
# 全テスト実行
test: test-go

# Go テスト
test-go:
    cd services/product-mgmt && go test ./...
    cd services/customer-mgmt && go test ./...
    cd services/ec-site && go test ./...

# --- Secret Scan ---
# シークレット検出 (gitleaks)
secret-scan:
    gitleaks detect --verbose

# --- Format ---
# 全 format 実行
fmt: fmt-ts fmt-go

# TypeScript format (oxfmt)
fmt-ts:
    oxfmt services/bff/src/
    oxfmt services/ec-site/frontend/src/

# Go format (golangci-lint fmt)
fmt-go:
    cd services/product-mgmt && golangci-lint fmt ./...
    cd services/customer-mgmt && golangci-lint fmt ./...
    cd services/ec-site && golangci-lint fmt ./...

# --- Format Check ---
# 全 format check
fmt-check: fmt-check-ts fmt-check-go

# TypeScript format check
fmt-check-ts:
    oxfmt --check services/bff/src/
    oxfmt --check services/ec-site/frontend/src/

# Go format check
fmt-check-go:
    cd services/product-mgmt && golangci-lint fmt --diff ./...
    cd services/customer-mgmt && golangci-lint fmt --diff ./...
    cd services/ec-site && golangci-lint fmt --diff ./...

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
    echo ""
    echo "=== bmkr dev services ==="
    echo "  product-mgmt:       http://localhost:8081"
    echo "  customer-mgmt:      http://localhost:8082"
    echo "  ec-site (backend):  http://localhost:8080"
    echo "  bff:                http://localhost:3000"
    echo "  ec-site (frontend): http://localhost:5173"
    echo "========================="
    echo ""
    (cd services/product-mgmt && go run .) &
    (cd services/customer-mgmt && go run .) &
    (cd services/ec-site && go run .) &
    (cd services/bff && pnpm dev) &
    (cd services/ec-site/frontend && pnpm dev) &
    wait

# Ory Hydra OAuth 2.0 クライアント登録
hydra-setup:
    bash scripts/hydra-setup.sh
