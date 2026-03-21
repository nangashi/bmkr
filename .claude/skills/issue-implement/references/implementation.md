# Phase 1: I/F設計 — サブエージェント詳細手順

計画の変更ファイル一覧をもとに、型・インターフェース・関数シグネチャを定義し、動作コメントを記述する。実装本体は書かない。

---

## プロンプトに含める情報

- 計画全文（実装ステップ・変更ファイル一覧・リスクと対策）
- Issue 本文（方針・受け入れ条件・スコープ外）
- 作業ディレクトリパス
- CLAUDE.md の内容
- 「実装本体は書かない。型・インターフェース・シグネチャ・動作コメントのみ」
- 出力先: `.output/issue-implement/{issue_number}/`

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

**TypeScript/React の場合:**
- Props 型の定義
- ページコンポーネントのシグネチャ
- ルーティング定義
- フロントエンドでは TODO プレースホルダーではなく実際の JSX 構造を記述すること。UI のコントラクト（要素構成・データバインディング・リンク先）がコードから読み取れる状態にする。データ取得ロジック（useEffect 内）は TODO コメントで良い

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

この動作コメントが Phase 3（テスト作成）の入力になる。テスト作成者はこのコメントと受け入れ条件からテストケースを導出する。

### 4. ビルド確認

I/F定義が完了したら `go build` / `tsc --noEmit` でコンパイルが通ることを確認する。空実装でもコンパイルは通るべき。

---

## 非自明な注意点

### 実装ガイドを読む

Go サービスの場合は `docs/guides/go/` 配下のガイドを全て読むこと。特に重要:
- `docs/guides/go/sqlc-type-boundary.md` — レイヤー間の型分離
- `docs/guides/go/handler-testability.md` — DB 依存のインターフェース注入
- `docs/guides/go/codegen-order.md` — 自動生成コードの実行順序

### 周辺の既存コードを必ず読む

変更対象ファイルの周辺にある既存ファイルを Read して、プロジェクト固有のパターンを把握してから定義する。

### スコープ外には触れない

Issue のスコープ外に記載された内容には触れない。

### 動作コメントは具体的に書く

「適切にエラーハンドリングする」ではなく「DB エラー時は echo.NewHTTPError(500) を返す」のように、テストケースを導出できる具体性で書く。
