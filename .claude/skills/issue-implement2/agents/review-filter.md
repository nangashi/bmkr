# レビュー指摘 採用判定サブエージェント

レビューで挙がった指摘を、実装コードとプロジェクトコンテキストだけを根拠に独立評価する。

## 入力パラメータ

- `issue_number`
- `review_output_path`
- `output_path`

---

## 前提情報の自己取得

1. `review_output_path` を Read してレビュー結果を把握する
2. `gh issue view {issue_number}` で Issue 本文を読む
3. `.output/issue-implement2/{issue_number}/oscillation-directives.md` が存在すれば Read する
4. `.output/issue-implement2/{issue_number}/contract.md` が存在すれば Read する

## 判定プロセス

### 1. コードの確認

指摘が参照するファイル・行を Read で確認する。レビュー結果の記述だけで判断しない。

### 2. 評価軸

| 軸 | 質問 |
|----|------|
| リスク | 対応しない場合に何が起きるか |
| 発生確率 | 通常利用で現実的に起きるか |
| 修正コスト | 修正範囲は Issue のスコープ内か |
| スコープ | 受け入れ条件と契約に照らして必要か |

### 3. 除外基準

- 到達しないリスク
- ADR やプロジェクト方針と不一致の提案
- コスト対効果が低い大規模リファクタリング
- スタイルの好み
- スコープ外
- oscillation directive と競合する提案

## 出力

`output_path` に Markdown ファイルを書き出す。

| finding_id | source | severity | decision | reason |
|------------|--------|----------|----------|--------|
