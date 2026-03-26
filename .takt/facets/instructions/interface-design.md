計画に基づき、型・インターフェース・関数シグネチャを定義し、動作コメントを記述する。実装本体は書かない。

`{report:init-context.md}` から Issue 情報と計画を読む。

## 参照資料の読み込み（必須）

1. `docs/guides/workflow/` 配下で関連するガイドを Read で読む
2. 変更対象ファイルの周辺にある既存ファイルを Read して、プロジェクト固有のパターンを把握する

## 自動生成コードの前提作業

計画に proto/SQL の変更が含まれる場合、先にこれらを実行する:
1. proto ファイルの変更 → `buf generate`
2. SQL クエリファイルの変更 → `sqlc generate`
3. 依存追加 → `go mod tidy` / `pnpm install`

## 型・インターフェースの定義

計画の各ステップで新規作成・変更するファイルについて、型・インターフェース・シグネチャを定義する。

## 関数シグネチャ + 動作コメント

シグネチャを書き、本体は空実装（`panic("not implemented")` や `return nil`）にする。
動作コメントには `// wip:` プレフィックスを付ける。

```go
// wip: HandleProductList handles GET /admin/products.
// wip: 動作:
// wip:   - page パラメータをパース、不正値は 1 にフォールバック
// wip:   - DB エラー時は echo.NewHTTPError(500) を返す
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    panic("not implemented")
}
```

動作コメントは Phase 3（テスト作成）の入力になるので、正常系だけでなく、分岐・エッジケース・異常系も考慮して網羅的に記載する。非機能制約（既存レスポンス構造の維持、DB 呼び出し回数の制限等）も含める。

## ビルド確認

I/F 定義が完了したら `go build ./...` / `tsc -b` でコンパイルが通ることを確認する。

## 計画差分の記録

I/F 設計で計画を具体化する際に、計画に書かれていないことを決めた箇所を Decisions レポートに記録する。

**記録する判断:**
- 計画に明記されていないが暗黙に決めたこと
- 計画の方針を変更・具体化したこと
- 複数の実装方法から特定の方法を選んだこと

**記録しない判断:**
- 既存コードのパターン踏襲
- `.claude/rules/` に従った選択

## スコープ外には触れない

Issue のスコープ外に記載された内容には触れない。
