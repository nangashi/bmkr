---
name: issue-implement
description: issue-plan で作成した実装計画をもとに、実装・テスト・レビュー・PR作成まで行う。`/issue-implement` で起動。
allowed-tools: Bash, Read, Glob, Grep, Agent, Write, Edit, AskUserQuestion, Skill
disable-model-invocation: true
argument-description: "Issue番号。例: '#123', '123'"
---

# issue-implement

GitHub Issue の実装計画をもとに、実装からPR作成まで自動で進める。issue-plan で作成された計画が Issue コメントに存在することが前提。

## 前提

- `/issue-plan` で計画が作成済みであること
- 計画は Issue コメントに `<!-- issue-plan:plan:done -->` マーカーで記録されている
- 実装不可の問題がなければ、PR作成まで対話なしで自動実行する

## ワークフロー

```
Phase 0: 初期化        → Issue・計画読み込み、ブランチ作成
Phase 1: I/F設計       → 型・インターフェース・シグネチャ定義 + 動作コメント
Phase 2: 設計レビュー   → Codex 汎用レビュー + プロジェクト固有レビュー
Phase 3: テスト作成     → I/F + 動作コメントからユニットテスト導出 — Red
Phase 4: 実装          → テスト Green + Codex Review + ゴール整合性
Phase 5: PR作成        → 受け入れ条件チェック・コミット・PR作成
```

このフローは I/F定義→テスト→実装の順（TDD）で進む。従来の「実装後にテスト」で起きていた問題（テストが実装の内部構造に依存する、設計問題の発見が遅れる）を解消するための設計。I/F定義フェーズで型・インターフェースが先に存在するため、TDD のコンパイルエラー問題も発生しない。

## Phase 0: 初期化

### 0a. 前提チェック

- `gh auth status` で GitHub CLI の利用可否を確認
- `which codex` で codex CLI の利用可否を確認（なければ Phase 2 の Codex でのレビューは Opus にフォールバック）

### 0b. 引数の解析

| パターン | 動作 |
|---------|------|
| `#123` または `123` | Issue を読み込む |
| 引数なし | `Issue番号を教えてください。` と表示して入力を待つ |

### 0c. 計画の検出

Issue コメントから `<!-- issue-plan:plan:done -->` マーカーを検索する。

計画が見つからない場合は以下を表示して終了する。

```
この Issue には実装計画がありません。
先に `/issue-plan #{番号}` で計画を作成してください。
```

### 0d. Issue の読み込み

Issue本文およびコメントから実装の前提情報を読み込む。

### 0e. worktree チェック

`git rev-parse --git-common-dir` と `git rev-parse --git-dir` を比較し、異なれば worktree 上で動作していると判定する。

worktree 上でなければ以下を表示して終了する:

```
このスキルは worktree 上で実行する必要があります。
`claude -w` でworktreeを作成してから再実行してください。
```

worktree 上であれば、`claude -w` で作成済みのブランチをそのまま使用して続行する。中間ファイル（レポート等）の出力先は worktree 内の `.output/issue-implement/{issue_number}/` とする。

---

## Phase 1: I/F設計

Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/interface-design.md` を読み込み、その手順に従って型・インターフェース・関数シグネチャ・動作コメントを定義する。

---

## Phase 2: 設計レビュー

I/F定義のみを対象にレビューを実施する。`references/design-review.md`に従い、Codex 汎用レビューとプロジェクト適合性レビュー（Sonnet サブエージェント、`model: "sonnet"`）を **並列実行** する。

両レビューの指摘を集約し、共通指摘は自動採用、片方のみの指摘はメインエージェントが妥当性を判断する。採用された指摘はメインエージェント（Opus）が直接 I/F定義を修正してから Phase 3 に進む。

---

## Phase 3: テスト作成 — Red

I/F定義の動作コメントをもとにユニットテストを導出する。実装が存在しないため、テストが実装に依存することは原理的にできない。

1. Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/test-derivation.md` を読み込んで作業する
2. サブエージェント完了後、`just test` を実行する。コンパイルが通り、テストが失敗する（Red）のが正しい状態。コンパイルエラーは I/F定義の問題なので修正する

---

## Phase 4: 実装

`references/implementation-loop.md` に従う。Codex CLI による実装 → テスト実行 → スコア判定 → 方針変更してリトライの最大3回で進める。3回で解決しなければ ESCALATE（テスト/I/F修正を試行して最終1回）。それでもダメなら中断してユーザーに報告する。全テスト通過後、最終ゲート（Codex Review → 採用判定 → Codex 修正を最大2回 + ゴール整合性レビュー）を経て Phase 5 へ。

---

## Phase 5: PR 作成

レビュー・修正が完了した実装を PR として提出する。`references/pr-creation.md` に従い、受け入れ条件チェック、コミット・プッシュ・PR 作成を行う。

