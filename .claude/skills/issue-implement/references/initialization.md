# 初期化

Issue・計画の読み込みとブランチ準備を行う。

---

## 0a. 前提チェック

- `git rev-parse --git-common-dir` と `git rev-parse --git-dir` を比較し、異なれば worktree 上で動作していると判定する。worktree 上でなければ以下を表示して終了する:
  ```
  このスキルは worktree 上で実行する必要があります。
  `claude -w` でworktreeを作成してから再実行してください。
  ```
  worktree 上であれば、`claude -w` で作成済みのブランチをそのまま使用して続行する。
- `gh auth status` で GitHub CLI の利用可否を確認
- `which codex` で codex CLI の利用可否を確認。なければ以下を表示して終了する:
  ```
  codex CLI が見つかりません。
  このスキルは codex CLI（設計レビュー・品質ゲートレビュー）を必須とします。
  codex CLI をインストールしてから再実行してください。
  ```

## 0b. 開発準備

worktree では gitignore 対象のファイル（node_modules, 生成コード）が存在しないため、以下を実行する:

1. `pnpm install`
2. `just generate`

## 0c. 引数の解析

| パターン | 動作 |
|---------|------|
| `#123` または `123` | Issue を読み込む |
| 引数なし | `Issue番号を教えてください。` と表示して入力を待つ |

## 0d. 計画の検出

Issue コメントから `<!-- issue-plan:plan:done -->` マーカーを検索する。

計画が見つからない場合は以下を表示して終了する。

```
この Issue には実装計画がありません。
先に `/issue-plan #{番号}` で計画を作成してください。
```

## 0e. Issue の読み込み

Issue本文およびコメントから実装の前提情報を読み込む。中間ファイル（レポート等）の出力先は worktree 内の `.output/issue-implement/{issue_number}/` とする。
