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
Phase 1: I/F設計       → 型・インターフェース・シグネチャ定義 + 動作コメント + 設計レビュー
Phase 2: テスト作成     → I/F + 動作コメントからユニットテスト導出 — Red
Phase 3: 実装          → テスト Green + 品質ゲート（Codex 3観点レビュー）
Phase 4: ゴール整合性   → 受け入れ条件トレーサビリティ・スコープ逸脱チェック
Phase 5: PR作成        → コミット・PR作成
Phase 6: 振り返り       → 空振り・手戻り検出、知見蓄積、改善提案
```

Phase 1〜3 は計画の内容に応じて省略・簡略化される。各 Phase の冒頭に記載された条件を参照すること。

このフローは I/F定義→テスト→実装の順（TDD）で進む。従来の「実装後にテスト」で起きていた問題（テストが実装の内部構造に依存する、設計問題の発見が遅れる）を解消するための設計。I/F定義フェーズで型・インターフェースが先に存在するため、TDD のコンパイルエラー問題も発生しない。

## Phase 0: 初期化

`references/initialization.md` に従い、前提チェック・開発準備・引数解析・計画検出・Issue 読み込みを行う。

---

## Phase 1: I/F設計

`references/interface-design-flow.md` に従い、計画の内容に基づいて実行方法を選択し、型・インターフェース・シグネチャ・動作コメントを定義する。

新しい型・インターフェースが定義された場合は、`references/interface-design-review.md` に従い設計レビュー（Codex レビュー → 採用判定 → 修正）を実施する。定義がなかった場合はスキップし、品質は Phase 3 の品質ゲートで担保する。

機械的な置換・削除のみの場合（動作コメントなし）は、編集後に `just test` + `just fmt` + `just lint` を実行して検証する。Phase 2・3 はスキップされ Phase 4 に進む。

---

## Phase 2: テスト作成 — Red

### スキップ判断

Phase 1 で `// wip:` 動作コメントが記述されなかった場合はスキップする。動作コメントがない = テスト対象の新規ロジックが存在しない。既存テストの通過で品質を担保する。

### 実施する場合

`// wip:` マーカー付きの動作コメントをもとにユニットテストを導出する。実装が存在しないため、テストが実装に依存することは原理的にできない。

1. Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/test-derivation.md` を読み込んで作業する
2. サブエージェント完了後、`just test` を実行する。コンパイルが通り、テストが失敗する（Red）のが正しい状態。コンパイルエラーは I/F定義の問題なので修正する

---

## Phase 3: 実装

Phase 2 がスキップされた場合は Phase 丸ごとスキップ。Phase 1 で編集・検証が完了しているため Phase 4 に進む。

`references/implementation.md` に従う。実装ループ（テスト Green） → wip クリーンアップ → 品質ゲート（ベースライン確認 + Codex 3観点レビュー + simplify + コメント整理）。

---

## Phase 4: ゴール整合性

常に実行する。`references/goal-alignment.md` に従い、受け入れ条件のトレーサビリティとスコープ逸脱をチェックする。

---

## Phase 5: PR 作成

Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/pr-creation.md` を読み込み、コミット・プッシュ・PR 作成を自律実行する。PR URL を受け取って表示する。

---

## Phase 6: 振り返り

Phase 5 完了後、または Phase 3 中断後に実行する。Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/retrospective.md` を読み込み、`.output/` 内のファイルを自己取得して知見の収集・原因分析・改善提案を行う。

メインエージェントはサブエージェントから受け取った改善提案をユーザーに提示し、承認された改善を実行するか改善 Issue を作成する。

