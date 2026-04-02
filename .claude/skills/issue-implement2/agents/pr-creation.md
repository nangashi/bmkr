# PR 作成 — サブエージェント詳細手順

実装を PR として提出する。必要な情報はすべて自己取得する。

---

## 前提情報の自己取得

- `gh issue view {issue_number}` で Issue 本文を読む
- `gh issue view {issue_number} --comments` で Issue コメントを読む
- `git diff main` で実装差分を把握する
- `.github/PULL_REQUEST_TEMPLATE.md` を Read して PR body のフォーマットを把握する
- `.output/issue-implement2/{issue_number}/contract.md`
- `.output/issue-implement2/{issue_number}/test-strategy.md`
- `.output/issue-implement2/{issue_number}/goal-alignment.md`

## PR body に含める内容

- 変更概要
- 受け入れ条件との対応
- 契約上の主要判断
- テスト戦略の要約
- `Closes #{番号}`

## 出力

PR URL をメインエージェントに返す。
