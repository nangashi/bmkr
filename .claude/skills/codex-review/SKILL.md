---
name: codex-review
description: Codex CLIを使ったコードレビューを実行する。ファイル指定・git diff指定・レビュー観点の自由テキスト指定に対応し、構造化された指摘を返す。指摘は対応合理性で自動フィルタリングされ、対応価値のあるものだけが残る。`/codex-review` やユーザーが「コードレビュー」「codexでレビュー」「レビューして」「diff見て」「変更をチェック」等と言った場合にトリガーする。codex CLIがインストールされている環境でのみ動作する。
allowed-tools: Bash, Read, Glob, Grep, Agent, Write, AskUserQuestion
user-invocable: true
argument-description: "レビュー対象と観点を指定。例: 'src/main.go セキュリティ観点で', '--diff HEAD~3', '--diff main..feature 保守性の観点で'"
---

# Codex Review

Codex CLI（`codex exec`）でコードレビューを実行し、対応合理性のある指摘だけを構造化して返すスキル。

**重要: ユーザーとの対話およびレビュー結果は日本語で出力すること。**

## ワークフロー概要

1. レビュー対象の特定（ファイル or git diff）
2. レビュー観点の確認
3. サブエージェントで codex exec 実行 + 対応合理性フィルタリング
4. フィルタ済み結果をユーザーに返却

## Step 1: 前提チェック

Bash で `which codex` を実行し、codex CLI が利用可能か確認する。利用できない場合は「codex CLI がインストールされていません」と伝えて終了する。

## Step 2: レビュー対象の特定

`$ARGUMENTS` を解析する:

### パターン A: ファイル指定
- ファイルパスが含まれている場合（例: `src/main.go src/utils.go`）
- Glob で対象ファイルの存在を確認する

### パターン B: git diff 指定
- `--diff` フラグがある場合（例: `--diff HEAD~3`, `--diff main..feature`）
- diff 指定文字列を抽出する

### パターン C: 引数なし or 曖昧
- ユーザーに確認する:

```
レビュー対象を教えてください:
1. ファイルパス（例: src/main.go）
2. git diff（例: --diff HEAD~3, --diff main..feature, --diff --staged）
```

## Step 3: レビュー観点の確認

`$ARGUMENTS` から観点を抽出する。観点は自由テキストで指定される（例: 「セキュリティ観点で」「パフォーマンスと保守性について」）。

- 観点が指定されている場合: そのまま使用する
- 観点が指定されていない場合: 汎用的なコードレビューとして実行する（観点を限定しない）

## Step 4: サブエージェントでレビュー実行 + フィルタリング

`.claude/skills/codex-review/agents/review-and-filter.md` を Read で読み込み、その内容をベースに Agent ツール（subagent_type=general-purpose）を起動する。

### サブエージェントに渡す情報:
- レビュー対象（ファイルパスのリスト or diff 指定文字列）
- レビュー対象の種別（file or diff）
- レビュー観点（指定がある場合）
- 作業ディレクトリのパス
- 出力ファイルパス: `.output/codex-review/YYYYMMDD-{review_target}.md`
  - `YYYYMMDD` は実行日（例: `20260320`）
  - `{review_target}` はレビュー対象を kebab-case で要約したもの（例: `backend-cmd-api-main`, `diff-HEAD-3`, `diff-main-feature`）

## Step 5: 結果の返却

サブエージェントの結果を受け取り、出力された `.output/codex-review/YYYYMMDD-{review_target}.md` を Read で読み込んで以下を提示する:

### 表示フォーマット:

出力ファイルの内容をそのまま提示する。出力ファイルには以下が含まれる:
- Summary（対象・観点・件数）
- Recommended Action Order（対応推奨順）
- Findings（各指摘の詳細 + コード例）
- Filtered Out（除外された指摘）

出力ファイルの末尾パスを案内する:
```
詳細: .output/codex-review/YYYYMMDD-{review_target}.md
```

指摘が0件の場合は「対応が必要な指摘はありませんでした」と伝える。
