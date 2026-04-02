# 契約設計

公開契約を先に定義する。ここで固定するのは公開 I/F、エラー契約、不変条件、生成コード前提、受け入れ条件対応表であり、内部 I/F はまだ凍結しない。

---

## 目的

- 契約を実装より先に固定する
- テスト導出の真実源泉を `wip` コメントではなく契約ファイルに寄せる
- 後続の実装で守るべき不変条件を明文化する

## 手順

1. `gh issue view {issue_number}` および `gh issue view {issue_number} --comments` で Issue 本文・コメントを読む
2. 変更対象ファイルの周辺コードを読み、既存の公開契約と内部責務の境界を把握する
3. `docs/adr/` と `docs/guides/codegen-workflow.md` を読み、技術的制約を確認する
4. `.output/issue-implement3/{issue_number}/contract.md` を作成する

## contract.md に必ず含める項目

```markdown
# Contract

## Public Interface
- 変更または新設する公開 I/F

## Error Contract
- どの条件で何を返すか

## Invariants
- 今回の変更で壊してはいけないこと

## Generated Code Preconditions
- proto/sqlc/templ などの前提

## Acceptance Mapping
| 受け入れ条件 | 契約上の対応箇所 |
|-------------|------------------|
```

## ルール

- 公開 API は具体的に書く
- エラー契約は「適切に処理する」と書かず、戻り値やレスポンスを明示する
- 不変条件にはレスポンス形状、ルーティング、主要副作用、スコープ外制約を含める
- 内部 I/F は必要最小限だけ書く。実装のための早すぎる分割は避ける

## 生成コードを含む場合

proto/SQL の変更を含む作業では以下の順序を守る:

1. proto 変更 → `buf generate`
2. SQL 変更 → `sqlc generate`
3. 依存追加 → `go mod tidy` / `pnpm install`
4. 手動コード

## 出力

- `.output/issue-implement3/{issue_number}/contract.md`
- `.output/issue-implement3/{issue_number}/contract-decisions.md`
- 必要に応じてスタブやシグネチャをコードに追加する

スタブを追加する場合は以下に従う:

- 実装本体は書かない
- `panic("not implemented")` などの未実装スタブに留める
- 既存コード内の変更は `// wip:` コメントで振る舞いと不変条件を補足してよい

## contract-decisions.md

計画（`<!-- issue-plan:plan:done -->` コメント内の `### 設計判断` テーブル）と、契約設計で実際に行った判断の差分を記録する。計画どおりの判断はスキップし、差分のみを書く。差分がなければ「計画どおり」と1行だけ書く。

```markdown
# Contract Decisions

| # | 計画の記述 | 契約での判断 | 変更理由 |
|---|-----------|-------------|---------|
```

このファイルは PR 作成時の `## Design Decisions` セクションと振り返りで参照される。
