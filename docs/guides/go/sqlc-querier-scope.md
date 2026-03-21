---
id: sqlc-querier-scope
category: testing
scope: [backend, admin-ui]
severity: medium
detectable_by_linter: false
---

# sqlc Querier インターフェースのスコープ管理

## アンチパターン

sqlc generate で Querier インターフェースにメソッドが追加されたとき、既存テストのモック実装が新メソッドを持たずコンパイルエラーになることに気づかない。

## 正しいパターン

SQL クエリを追加して `sqlc generate` を実行した後、既存テストのモック実装にもメソッドスタブを追加する。これは I/F 設計フェーズの自動生成コード前提作業の一部として行う。

## 確認手順

1. `sqlc generate` を実行
2. Querier インターフェースの diff を確認
3. 既存テストで Querier をモックしている箇所を検索
4. 新メソッドのスタブを追加（`panic("not called")` で良い）
5. `go build ./...` で確認
