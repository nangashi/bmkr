# ゴール整合性

実装結果が Issue の受け入れ条件を満たしているか、スコープを逸脱していないかを検証する。常に実行する。

---

## 実行方法

Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/goal-alignment.md` を読み込んで作業する。

## 結果の処理

サブエージェントの出力（`.output/issue-implement2/{issue_number}/goal-alignment.md`）を確認する。

- 全て「OK」→ Phase 9 へ
- 「未対応」または「要確認」がある → ユーザーに報告し、対応を判断する
