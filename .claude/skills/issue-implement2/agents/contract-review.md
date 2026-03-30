# 契約レビュー — サブエージェント詳細手順

公開契約、エラー契約、不変条件、受け入れ条件対応表を独立コンテキストでレビューする。

---

## 前提情報の取得

- `gh issue view {issue_number}` と `gh issue view {issue_number} --comments` で Issue 本文・コメントを読む
- `.output/issue-implement2/{issue_number}/contract.md` を読む
- 必要に応じて `docs/adr/` と `docs/guides/` を読む

## 見るべき観点

- 公開 I/F が過不足なく定義されているか
- エラー契約が曖昧でないか
- 不変条件が明示されているか
- 生成コード前提が漏れていないか
- 受け入れ条件と契約の対応が取れているか

## 注意

- 実装案は考えなくてよい
- 内部 I/F を増やす提案は慎重に扱う
- スコープ外の改善は指摘しない
