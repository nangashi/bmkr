# PR 作成（Phase 5）

レビュー・修正が完了した実装を PR として提出する。

---

## 5a. 受け入れ条件の確認 & チェックボックス更新

Issue 本文の受け入れ条件を1つずつ確認する。満たされている条件のチェックボックスを `- [x]` に更新する。

**注意:** `gh issue edit --body` にインラインで本文を渡すと特殊文字でエスケープ問題が起きる。一時ファイルに書き出して `--body-file` を使うこと。

満たされていない条件がある場合は、その旨をユーザーに報告し、PR の Review Notes にも記載する。

---

## 5b. Design Decisions セクションの生成

PR body の `## Design Decisions` セクションを以下の情報源から構成する。

**情報源:**

1. Issue コメント（`<!-- issue-plan:approach:done -->`）の `### 設計判断` テーブル
2. `.output/issue-implement/{issue_number}/interface-design-decisions.md`

**手順:**

1. `gh issue view {issue_number} --comments` から方針検討コメントの設計判断テーブルを抽出する（テーブルがなければスキップ）
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

## 5c. コミット・プッシュ・PR 作成

`.github/PULL_REQUEST_TEMPLATE.md` のフォーマットに従って PR body を作成する。`Closes #{番号}` を含めること。`## Design Decisions` セクションは 5b で生成した内容を使用する。

---

## 5d. 完了報告

PR URL を表示する。

