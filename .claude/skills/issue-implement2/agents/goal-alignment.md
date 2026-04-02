# ゴール整合性レビュー — サブエージェント詳細手順

実装結果が Issue の受け入れ条件を満たしているか、スコープを逸脱していないかを検証する。

---

## 前提情報の取得

- `gh issue view {issue_number}` で Issue 本文を読む
- `gh issue view {issue_number} --comments` で計画コメントを読む
- `.output/issue-implement2/{issue_number}/contract.md` を読む
- `git diff main` で実装差分を把握する

## トレーサビリティ表の作成

受け入れ条件ごとに、対応するコードとテストを特定し、以下の表を埋める:

| 受け入れ条件 | 対応コード | 対応テスト | 状態 |
|-------------|-----------|-----------|------|

## スコープ逸脱チェック

- Issue のスコープ外に触れていないか
- 計画にない追加機能・大規模リファクタリングが紛れていないか
- 契約の独自解釈がないか

## 出力

結果を `.output/issue-implement2/{issue_number}/goal-alignment.md` に書き出す。
