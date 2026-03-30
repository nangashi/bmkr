# Red レビュー

Red 生成後に、契約追従か、粒度が適切か、実装依存になっていないかをレビューする。

---

## Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff HEAD`
- `{perspective_files}`: `testability.md`
- `{output_path}`: `.output/issue-implement2/{issue_number}/review-red.md`

## 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:

- `issue_number`
- `review_output_path`（`.output/issue-implement2/{issue_number}/review-red.md`）
- `output_path`（`.output/issue-implement2/{issue_number}/review-red-filtered.md`）

## 判定結果の処理

- 採用指摘あり → Opus がテストや契約を修正する
- 採用指摘なし → Phase 4 に進む
