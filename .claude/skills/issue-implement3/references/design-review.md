# 設計レビュー

L2 の簡易設計レビューおよび L3 の契約レビューで使用する共通手順。

---

## Step 1: Codex レビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff HEAD`（設計 Phase で追加された変更のみ。ブランチ全体の diff ではない）
- `{perspective_files}`: `type-design.md`, `error-contract.md`, `testability.md`
- `{output_path}`: `.output/issue-implement3/{issue_number}/review-design.md`

### L3 の場合: 契約情報のインライン化

perspectives ファイルに contract.md の以下のセクションをインラインで埋め込む:

- Public Interface セクション → type-design.md に追加
- Error Contract セクション → error-contract.md に追加
- Invariants セクション → testability.md に追加

これにより Codex は外部ファイルを参照せず、perspectives 内の情報だけでレビューを完結できる。

### L2 の場合

perspectives のみ（インライン化なし）。自己完結型の perspectives で十分。

---

## Step 2: 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:
- `issue_number`
- `review_output_path`（`.output/issue-implement3/{issue_number}/review-design.md`）
- `output_path`（`.output/issue-implement3/{issue_number}/review-design-filtered.md`）

4 軸で判定する:

- **正確性**: 指摘はコードの事実に基づいているか
- **重大度**: CRITICAL / IMPORTANT / LOW
- **スコープ**: Issue の受け入れ条件と計画の範囲内か
- **費用対効果**: 修正コストに見合う効果があるか

CRITICAL / IMPORTANT かつスコープ内の指摘のみ採用する。

---

## Step 3: 修正

採用された指摘を Opus が修正する（設計/契約の修正は判断を伴うため Opus）。

修正後、ビルド確認（`just build` or コンパイル）を実行する。
