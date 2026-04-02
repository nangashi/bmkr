# 契約設計 — サブエージェント詳細手順

計画の変更ファイル一覧をもとに、公開 I/F・エラー契約・不変条件を定義し、contract.md と contract-decisions.md を出力する。実装本体は書かない。

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

生成コードが存在しないと後続の定義でインポートエラーになるため、この順序は必須。

### 2. contract.md の作成

`references/contract-design.md` のフォーマットに従い、以下を記録する:

- **Public Interface**: 変更・新設する公開 I/F（具体的なシグネチャ付き）
- **Error Contract**: どの条件で何を返すか（「適切に処理する」は不可）
- **Invariants**: 今回の変更で壊してはいけないこと
- **Generated Code Preconditions**: proto/sqlc/templ の前提
- **Acceptance Mapping**: 受け入れ条件と契約箇所の対応表

出力先: `.output/issue-implement2/{issue_number}/contract.md`

### 3. 型・インターフェースの定義

計画の各ステップで新規作成・変更するファイルについて、以下を定義する:

**Go の場合:**
```go
// adminQuerier は管理画面ハンドラが必要とする DB 操作を定義する。
// *db.Queries がこのインターフェースを満たす。
type adminQuerier interface {
    ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error)
    CountProducts(ctx context.Context) (int64, error)
}
```

### 4. 関数・メソッドのシグネチャ + 動作コメント

シグネチャを書き、本体は空実装（`panic("not implemented")` や `return nil`）にする。動作コメントで正常系・異常系の振る舞いを記述する。

```go
// wip: HandleProductList handles GET /admin/products.
// wip:
// wip: 動作:
// wip:   - クエリパラメータ page（デフォルト: 1）をパース。不正値は 1 にフォールバック
// wip:   - CountProducts と ListProducts を errgroup で並列実行
// wip:
// wip: エラー:
// wip:   - DB エラー時は echo.NewHTTPError(500) を返す
// wip:   - page パラメータの不正値はエラーにせずデフォルト値にフォールバック
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    panic("not implemented")
}
```

動作コメントには `// wip:` プレフィックスを付けること。このマーカーによりテスト導出フェーズが既存コメントと動作コメントを区別でき、実装フェーズのクリーンアップ対象を特定できる。

動作コメントは正常系だけでなく、分岐・エッジケース・異常系も網羅的に記載する。テスト導出の入力になるため、テストケースを導出できる具体性で書く（「適切にエラーハンドリングする」は不可）。

### 5. ビルド確認

定義完了後に `go build ./...` / `tsc -b` でコンパイルが通ることを確認する。空実装でもコンパイルは通るべき。

**注意:** TypeScript では `tsc --noEmit` ではなく `tsc -b` を使うこと。`tsc --noEmit` は Project References 経由の設定を見逃す。

### 6. contract-decisions.md の作成

契約設計で計画から逸脱した判断を記録する。

**手順:**

1. 方針検討コメント（`<!-- issue-plan:approach:done -->`）の `### 設計判断` テーブルがあれば読み、そこに記載済みの判断は差分に含めない
2. Step 1-5 で変更・作成した各ファイルの内容を計画と突き合わせ、差分を抽出する
3. `.output/issue-implement2/{issue_number}/contract-decisions.md` に書き出す

**記録する判断:**
- 計画に明記されていないが暗黙に決めたこと（例: インターフェースに含めるメソッドの選定）
- 計画の方針を変更・具体化したこと（例: エラー契約の解釈を変えた）
- 複数の実装方法から特定の方法を選んだこと

**記録しない判断:**
- 既存コードのパターン踏襲（理由が自明）
- `.claude/rules/` に従った選択（ルール参照で十分）
- 方針検討の設計判断テーブルに記載済みの判断

**出力フォーマット:**

差分がある場合:
```markdown
# Contract Decisions

| # | 計画の記述 | 契約での判断 | 変更理由 |
|---|-----------|-------------|---------|
| 1 | {計画の該当箇所} | {契約設計で決めたこと} | {なぜその選択か} |
```

計画通りに設計した場合:
```markdown
# Contract Decisions

計画どおり。差分なし。
```

---

## 非自明な注意点

### 実装ガイドを読む

`docs/guides/codegen-workflow.md` を読む。

### 周辺の既存コードを必ず読む

変更対象ファイルの周辺にある既存ファイルを Read して、プロジェクト固有のパターンを把握してから定義する。

### 内部 I/F は凍結しない

公開 I/F・エラー契約・不変条件は固定するが、内部の実装構造を早期に決めない。内部 I/F は実装 Phase で最小限の変更を許容する。

### スコープ外には触れない

Issue のスコープ外に記載された内容には触れない。

### 非機能制約も動作コメントに含める

Issue 本文や計画に記載された非機能制約（既存レスポンス構造の維持、DB 呼び出し回数の制限等）も動作コメントに明記する。

---

## ブロッカー判定

以下の場合は作業を中断し、メインエージェントに報告する:
- 計画の前提が現在のコードと合わない
- 外部サービスの設定変更が必要
- 計画の曖昧さにより複数の解釈が可能
