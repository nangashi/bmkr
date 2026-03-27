# ゴール整合性

実装結果が Issue の受け入れ条件を満たしているか、スコープを逸脱していないかを検証する。常に実行する。

---

## 実行方法

Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/goal-alignment.md` を読み込んで作業する。構造化されたチェックリスト照合タスクのため Sonnet で十分な精度が得られる。

---

## 結果の処理

サブエージェントの出力（`.output/issue-implement/{issue_number}/goal-alignment.md`）を確認する。

- 全て「OK」→ Phase 5 へ
- 「未対応」または「要確認」がある → ユーザーに報告し、対応を判断する
