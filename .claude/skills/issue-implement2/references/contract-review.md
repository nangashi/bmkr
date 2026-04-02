# 契約レビュー

契約設計後に、Codex の独立レビューと Sonnet の採用判定を行う。

---

## Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff HEAD`
- `{perspective_files}`: `type-design.md`, `error-contract.md`, `testability.md`
- `{output_path}`: `.output/issue-implement2/{issue_number}/review-contract.md`

## 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:

- `issue_number`
- `review_output_path`（`.output/issue-implement2/{issue_number}/review-contract.md`）
- `output_path`（`.output/issue-implement2/{issue_number}/review-contract-filtered.md`）

## 判定結果の処理

- 採用指摘あり → Opus が契約とスタブを修正する
- 採用指摘なし → Phase 2 に進む

修正後は `just fmt` を実行して整形を確認する。
