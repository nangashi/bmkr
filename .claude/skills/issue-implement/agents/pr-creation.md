# PR 作成 — サブエージェント詳細手順

実装を PR として提出する。必要な情報はすべて自己取得する。

---

## 前提情報の自己取得

- `gh issue view {issue_number}` で Issue 本文を読む
- `gh issue view {issue_number} --comments` で Issue コメント（計画・方針検討）を読む
- `git diff main` で実装差分を把握する
- `.github/PULL_REQUEST_TEMPLATE.md` を Read して PR body のフォーマットを把握する

---

## Design Decisions セクションの生成

PR body の `## Design Decisions` セクションを以下の情報源から構成する。

**情報源:**

1. Issue コメント（`<!-- issue-plan:approach:done -->`）の `### 設計判断` テーブル
2. `.output/issue-implement/{issue_number}/interface-design-decisions.md`

**手順:**

1. Issue コメントから方針検討コメントの設計判断テーブルを抽出する（テーブルがなければスキップ）
2. `.output/issue-implement/{issue_number}/interface-design-decisions.md` を Read する（ファイルが存在しない場合、または「差分なし」の場合はスキップ）
3. テーブルの内容は原文のまま転記する。カラム名・行の文言を変更しない

**両方に内容がある場合:**

```markdown
## Design Decisions

### 方針検討での判断

{設計判断テーブルをそのまま転記}

### I/F設計での判断差分

{計画差分テーブルをそのまま転記}
```

**片方のみの場合:**

該当するサブセクションのみ出力する。

**両方が空の場合:**

`## Design Decisions` セクション自体を省略する。

---

## コミット・プッシュ・PR 作成

`.github/PULL_REQUEST_TEMPLATE.md` のフォーマットに従って PR body を作成する。`Closes #{番号}` を含めること。`## Design Decisions` セクションは上記で生成した内容を使用する。

---

## 出力

PR URL をメインエージェントに返す。
