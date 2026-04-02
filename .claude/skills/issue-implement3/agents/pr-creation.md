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

**情報源（存在するものだけ使用）:**

1. Issue コメント（`<!-- issue-plan:approach:done -->`）の `### 設計判断` テーブル
2. `.output/issue-implement3/{issue_number}/interface-design-decisions.md`（L2 の場合）
3. `.output/issue-implement3/{issue_number}/contract-decisions.md`（L3 の場合）
4. `.output/issue-implement3/{issue_number}/test-strategy.md`（L3 の場合、サマリーのみ）

**手順:**

1. Issue コメントから方針検討コメントの設計判断テーブルを抽出する（テーブルがなければスキップ）
2. `interface-design-decisions.md` または `contract-decisions.md` を Read する（ファイルが存在しない場合、または「差分なし」の場合はスキップ）
3. `test-strategy.md` が存在する場合、テスト戦略のサマリー（UT/統合/不要の分類一覧）を抽出する
4. テーブルの内容は原文のまま転記する。カラム名・行の文言を変更しない

**両方に内容がある場合:**

```markdown
## Design Decisions

### 方針検討での判断

{設計判断テーブルをそのまま転記}

### I/F設計・契約設計での判断差分

{計画差分テーブルをそのまま転記}

### テスト戦略（L3 の場合のみ）

{テスト戦略サマリー}
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
