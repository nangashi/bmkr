# Phase 1: I/F設計 — サブエージェント詳細手順

計画の変更ファイル一覧をもとに、型・インターフェース・関数シグネチャを定義し、動作コメントを記述する。実装本体は書かない。

---

## 前提情報の取得

- `gh issue view {issue_number}` および `gh issue view {issue_number} --comments` で Issue 本文・コメント（計画含む）を直接読む
- 実装本体は書かない。型・インターフェース・シグネチャ・動作コメントのみ

---

## 作業の流れ

### 1. 自動生成コードの前提作業

計画に proto/SQL の変更が含まれる場合、先にこれらを実行する:
1. proto ファイルの変更 → `buf generate`
2. SQL クエリファイルの変更 → `sqlc generate`
3. 依存追加（go.mod, package.json）→ `go mod tidy` / `pnpm install`

生成コードが存在しないと後続のI/F定義でインポートエラーになるため、この順序は必須。

### 2. 型・インターフェースの定義

計画の各ステップで新規作成・変更するファイルについて、以下を定義する:

**Go の場合:**
```go
// adminQuerier は管理画面ハンドラが必要とする DB 操作を定義する。
// *db.Queries がこのインターフェースを満たす。
type adminQuerier interface {
    ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error)
    CountProducts(ctx context.Context) (int64, error)
}

// ProductItem はテンプレートに渡す商品データ。
// db.Product を直接渡さず、プレゼンテーション用に変換した型を使う。
type ProductItem struct {
    ID            int64
    Name          string
    Price         int64
    StockQuantity int32
    CreatedAt     time.Time
}
```

### 3. 関数・メソッドのシグネチャ + 動作コメント

シグネチャを書き、本体は空実装（`panic("not implemented")` や `return nil`）にする。動作コメントで正常系・異常系の振る舞いを記述する。

```go
// HandleProductList handles GET /admin/products.
//
// 動作:
//   - クエリパラメータ page（デフォルト: 1）をパース。不正値は 1 にフォールバック
//   - CountProducts と ListProducts を errgroup で並列実行
//   - 総ページ数を計算。page が超過したら最終ページにクランプし再フェッチ
//   - HX-Request ヘッダの有無で分岐: HTMX → テーブルパーシャル、通常 → フルページ
//
// エラー:
//   - DB エラー時は echo.NewHTTPError(500) を返す
//   - page パラメータの不正値はエラーにせずデフォルト値にフォールバック
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    panic("not implemented")
}
```

この動作コメントが Phase 3（テスト作成）の入力になるので、正常系だけでなく、分岐・エッジケース・異常系も考慮して網羅的に記載する。テスト作成者はこのコメントと受け入れ条件からテストケースを導出する。

この動作コメントは実装完了後に最終ゲートでドキュメントコメントに変換される。Phase 1 時点ではテスト導出に必要な網羅性を優先する。

### 4. ビルド確認

I/F定義が完了したら `go build ./...` / `tsc -b` でコンパイルが通ることを確認する。空実装でもコンパイルは通るべき。

**注意:** TypeScript では `tsc --noEmit` ではなく `tsc -b` を使うこと。`tsc --noEmit` はルートの `tsconfig.json` のみを見るため、Project References 経由の `tsconfig.app.json`（`noUnusedLocals: true` 等）が適用されず、未使用インポートなどのエラーを見逃す。

---

## 非自明な注意点

### 実装ガイドを読む

`docs/guides/` 配下のガイドを全て読むこと。

### 周辺の既存コードを必ず読む

変更対象ファイルの周辺にある既存ファイルを Read して、プロジェクト固有のパターンを把握してから定義する。

### スコープ外には触れない

Issue のスコープ外に記載された内容には触れない。

### 動作コメントは具体的に書く

「適切にエラーハンドリングする」ではなく「DB エラー時は echo.NewHTTPError(500) を返す」のように、テストケースを導出できる具体性で書く。

### 非機能制約も動作コメントに含める

Issue 本文や計画に記載された非機能制約（既存レスポンス構造の維持、DB 呼び出し回数の制限、パフォーマンス要件等）も動作コメントに明記する。動作コメントに落ちなかった制約は Phase 3 のテストでカバーされず、退行しても検出できなくなる。

---

## ブロッカー判定

以下の場合は実装を中断し、ユーザーに報告する:
- 計画の前提が現在のコードと合わない
- 外部サービスの設定変更が必要
- 計画の曖昧さにより複数の解釈が可能
