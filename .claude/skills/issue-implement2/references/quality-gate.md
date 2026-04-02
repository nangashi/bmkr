# 品質ゲート

cluster 実装完了後に、横断品質を確認する。

---

## ベースライン確認

`just test` + `just fmt` + `just lint` を実行する。失敗時はエラーを分析して修正する。3回で解決しない場合は中断する。

## レビュー・修正ループ

```
git add -A && git commit -m "checkpoint: before quality gate (issue #{number})"

for round in 1..3:
  Codex 観点付きレビュー
  Sonnet 採用判定
  採用指摘なし → ループ終了

  Sonnet 修正
  git add -A && git commit -m "checkpoint: quality gate round {round} (issue #{number})"

  round >= 2 の場合:
    振動検出

simplify
コメント整理
```

## Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す。

- `{diff_command}`: `git diff main`
- `{perspective_files}`: `silent-failure.md`, `project-compliance.md`, `security.md`
- `{output_path}`: `.output/issue-implement2/{issue_number}/review-final.md`

## 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:

- `issue_number`
- `review_output_path`（`.output/issue-implement2/{issue_number}/review-final.md`）
- `output_path`（`.output/issue-implement2/{issue_number}/review-final-filtered.md`）

## 修正

採用された指摘は Sonnet サブエージェントで修正する。修正後に `just test`、`just fmt`、`just lint` を実行する。

## 振動検出

Round 2 以降、直前ラウンドの変更を打ち消す変更がないか確認する。検出した場合は `.output/issue-implement2/{issue_number}/oscillation-directives.md` に directive を追記する。

## simplify

レビュー対応で増えた冗長コードを整理する。`/simplify` を Skill ツールで実行してよい。

## コメント整理

`agents/comment-cleanup.md` を Read で読み込んだ Haiku モデルのサブエージェント（`model: haiku`）を起動し、`// 動作:` / `// エラー:` 形式のコメントをドキュメントコメントに変換する。
