---
id: codegen-workflow
category: workflow
scope: [backend, admin-ui]
severity: high
detectable_by_linter: false
---

# コード生成ワークフロー

proto/SQL の変更を含む作業で守るべき手順。

## 自動生成コードの実行順序

### アンチパターン

proto/SQL の変更を含む作業で、手動実装コードを先に書く。生成コードに依存する型・関数が未定義でコンパイルエラーになる。

### 正しいパターン

以下の順序を守る:

1. proto ファイルの変更 → `buf generate`
2. SQL クエリファイルの変更 → `sqlc generate`
3. 依存追加 → `go mod tidy` / `pnpm install`
4. 手動実装コード（型定義、ハンドラ等）

生成コードが存在してから手動コードを書くことで、import エラーや型不整合を防ぐ。

## sqlc 生成後のモック同期

### アンチパターン

sqlc generate で Querier インターフェースにメソッドが追加されたとき、既存テストのモック実装が新メソッドを持たずコンパイルエラーになることに気づかない。

### 正しいパターン

SQL クエリを追加して `sqlc generate` を実行した後、既存テストのモック実装にもメソッドスタブを追加する。これは I/F 設計フェーズの自動生成コード前提作業の一部として行う。

### 確認手順

1. `sqlc generate` を実行
2. Querier インターフェースの diff を確認
3. 既存テストで Querier をモックしている箇所を検索
4. 新メソッドのスタブを追加（`panic("not called")` で良い）
5. `go build ./...` で確認
