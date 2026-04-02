# テスト戦略レビュー — サブエージェント詳細手順

契約と変更内容から、テスト戦略の分類が妥当かを独立コンテキストで判定する。

---

## 前提情報の取得

- `.output/issue-implement2/{issue_number}/contract.md`
- `.output/issue-implement2/{issue_number}/test-strategy.md`
- `docs/guides/testing-strategy.md`

## 見るべき観点

- UT必須の箇所を落としていないか
- モックコストが高い箇所に無理に UT を割り当てていないか
- テスト不要判断が妥当か
- 受け入れ条件に対応する最低限の検証が存在するか
