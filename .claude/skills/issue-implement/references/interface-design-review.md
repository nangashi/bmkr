# 設計レビュー

I/F定義のみを対象に、Codex CLI で 3 観点（型設計・エラー契約・テスト駆動性）のレビューを実施し、採用判定を経て修正する。

---

## Codex 観点付きレビュー（Codex CLI）

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff HEAD`
- `{perspective_files}`: `type-design.md`, `error-contract.md`, `testability.md`
- `{output_path}`: `.output/issue-implement/{issue_number}/review-design.md`

---

## 採用判定（別サブエージェント）

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。レビュアとは独立したコンテキストで、各指摘をコードの事実に基づいて判断する。4軸評価と除外基準の適用を正確に行うため Sonnet を使用する。

渡すパラメータ: `issue_number`、`review_output_path`（`.output/issue-implement/{issue_number}/review-design.md`）、`output_path`（`.output/issue-implement/{issue_number}/review-design-filtered.md`）。Issue 本文や oscillation directives はサブエージェントが自己取得する。

### 判定結果の処理

- 採用指摘あり → 修正へ
- 採用指摘なし → 完了

---

## 修正（メインエージェント）

採用された指摘をメインエージェント（Opus）が直接 I/F定義を修正する。

修正後に `just fmt` を実行して整形を確認する。
