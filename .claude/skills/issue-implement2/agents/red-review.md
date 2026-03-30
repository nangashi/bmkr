# Red レビュー — サブエージェント詳細手順

導出したテストまたは検証項目が契約追従になっているかを独立コンテキストで確認する。

---

## 前提情報の取得

- `.output/issue-implement2/{issue_number}/contract.md`
- `.output/issue-implement2/{issue_number}/test-strategy.md`
- `.output/issue-implement2/{issue_number}/red-summary.md`
- 追加したテストコード

## 見るべき観点

- 実装内部に依存した期待値になっていないか
- 異なる振る舞いを1つのテストに詰め込みすぎていないか
- 境界条件と異常系が十分に具体的か
- テスト不要とした項目が本当に不要か
