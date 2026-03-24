---
globs: services/**/*_test.go
---

# Go テスト作成ルール

## 1. モックコスト上限

モック基盤（型定義 + フィクスチャビルダー）の行数がアサーション行数を超える場合、テスト手段を再検討する。

代替手段:
- `go/parser` や `grep` による静的解析テスト（import 検証、命名規約検証など）
- 既存テスト + lint/build の通過で担保

## 2. 依存の注入方式に応じたモック戦略

- **インターフェース DI 済み**（例: `AdminProductStore`）→ そのインターフェースをモック。`services/product-mgmt/admin_handler_test.go` を参照
- **`*db.Queries` 直接依存** → `db.DBTX` レベルのモック（`mockDBTX`, `mockRow`, `mockRows` 等）を新規作成しない。Scan のフィールド順序に依存するモックは sqlc 生成コードの内部構造に密結合し、壊れやすい
- **Connect `AnyRequest`/`AnyResponse` に依存する interceptor** → `httptest.NewServer` + `Unimplemented*Handler` 埋め込み + 生成クライアント経由でテストする。`AnyRequest`/`AnyResponse` は非公開メソッドを含み外部パッケージからモック実装できない。`services/ec-site/logging_interceptor_test.go` を参照

`*db.Queries` に直接依存するハンドラのテストが必要な場合は、先にインターフェースを抽出するか、静的解析テストで代替する。

## 3. 変更量との比例

変更が数行のリファクタリング（import 置換、rename、ログレベル変更等）の場合、数百行のテスト基盤を新規構築しない。まず静的解析テストで受け入れ条件をカバーできるか検討する。

## 4. テストの重複抑制

同一コードパスの異なるアサーションは 1 テストにまとめる。バリエーションはテーブルドリブンテストで表現する。

## 5. templ の出力に注意

templ は `<!doctype html>`（小文字）を出力する。テストで `<!DOCTYPE html>`（大文字）を期待値として書くと失敗する。templ 生成コードの実際の出力を確認してからテストの期待値を書くこと。
