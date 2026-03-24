---
id: codegen-order
category: workflow
scope: [backend, admin-ui]
severity: high
detectable_by_linter: false
---

# 自動生成コードの実行順序

## アンチパターン

proto/SQL の変更を含む作業で、手動実装コードを先に書く。生成コードに依存する型・関数が未定義でコンパイルエラーになる。

## 正しいパターン

以下の順序を守る:

1. proto ファイルの変更 → `buf generate`
2. SQL クエリファイルの変更 → `sqlc generate`
3. 依存追加 → `go mod tidy` / `pnpm install`
4. 手動実装コード（型定義、ハンドラ等）

生成コードが存在してから手動コードを書くことで、import エラーや型不整合を防ぐ。
