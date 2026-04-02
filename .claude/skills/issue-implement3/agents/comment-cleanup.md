# コメント整理 — サブエージェント詳細手順

`// 動作:` / `// エラー:` 形式の動作コメントを、実装完了後のコードに適したドキュメントコメントに変換する。

---

## 対象

`// 動作:` / `// エラー:` 形式のコメントを持つ関数・メソッド。`grep -rn "// 動作:\|// エラー:" --include="*.go" .` で検出する。

---

## 変換先の形式

```
// {サマリ: 何をするか 1-2文}
//
//   - {エッジケース、エラー、制約、戻り値の要点}
//   - {非自明な選択には括弧で理由を添える}
```

---

## 変換ルール

| 動作コメントの内容 | 変換先 |
|------------------|--------|
| メソッドの目的 | サマリ（1-2文） |
| 主要な振る舞いの分岐 | 重要ならサマリに含める |
| エッジケースの扱い | 箇条書き |
| エラーの種類と条件 | 箇条書き |
| 制約・前提条件 | 箇条書き or サマリ |
| 非自明な選択の理由 | 該当箇条書きに括弧で添える |
| 実装手順（HOW） | **削除** |
| コードから自明な WHAT | **削除** |

箇条書きが 0 件ならサマリのみ。自明なメソッドには箇条書き不要。

---

## 変換例

### 例1: 複数の分岐・エッジケースがある関数

Before:
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
```

After:
```go
// HandleProductList handles GET /admin/products.
// ページネーション付きで商品一覧を返す。HX-Request の有無でパーシャル/フルページを切り替える。
//
//   - page パラメータの不正値は 1 にフォールバック（エラーにしない）
//   - page が総ページ数を超過した場合は最終ページにクランプして再フェッチ
//   - DB エラー時は 500 を返す
func (h *AdminHandler) HandleProductList(c echo.Context) error {
```

### 例2: 自明なメソッド

Before:
```go
// ToProductItem converts db.Product to ProductItem.
//
// 動作:
//   - db.Product の各フィールドを ProductItem にマッピングする
//   - CreatedAt は UTC のまま変換する
func ToProductItem(p db.Product) ProductItem {
```

After:
```go
// ToProductItem converts db.Product to ProductItem.
func ToProductItem(p db.Product) ProductItem {
```

---

## 判断基準

- **「コードから自明」の判断**: 関数本体を読んで、コメントなしでも振る舞いが明確に分かる内容は削除する。フィールドの単純マッピング、標準的なエラー伝播などが該当する
- **残すべき内容**: フォールバック動作、暗黙の制約、非自明なエラー条件、パフォーマンス上の理由がある設計判断

---

## 検証

変換後に `just fmt` を実行して整形を確認する。
