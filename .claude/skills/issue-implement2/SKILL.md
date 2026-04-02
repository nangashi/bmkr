---
name: issue-implement2
description: issue-plan で作成した実装計画をもとに、契約設計、テスト戦略、failure cluster 単位の実装、品質ゲート、PR作成まで行う改善版スキル。`/issue-implement2` で起動。
allowed-tools: Bash, Read, Glob, Grep, Agent, Write, Edit, AskUserQuestion, Skill
disable-model-invocation: true
argument-description: "Issue番号。例: '#123', '123'"
---

# issue-implement2

GitHub Issue の実装計画をもとに、公開契約を先に固め、failure cluster 単位で実装を進める。`issue-plan` で作成された計画が Issue コメントに存在することが前提。

## 前提

- `/issue-plan` で計画が作成済みであること
- 計画は Issue コメントに `<!-- issue-plan:plan:done -->` マーカーで記録されている
- 実装不可の問題がなければ、PR作成まで対話なしで自動実行する
- メインのオーケストレーションは Opus が担当する

## フロー

```
Phase 0: 初期化          → Issue・計画読み込み、ブランチ準備、作業ディレクトリ準備
Phase 1: 契約設計        → 公開 I/F、エラー契約、不変条件、受け入れ条件を定義
Phase 2: テスト戦略      → UT / 統合 / テスト不要を判定
Phase 3: Red 生成        → 契約からテストまたは検証項目を導出
Phase 4: 失敗分類        → failure cluster に分解し、修正順を決める
Phase 5: 実装            → cluster 単位で Sonnet 実装 + 改善判定
Phase 6: Escalation      → 契約/テスト/実装の誤診を是正
Phase 7: 品質ゲート      → test/fmt/lint + Codex レビュー + simplify + コメント整理
Phase 8: ゴール整合性    → 受け入れ条件トレーサビリティ・スコープ逸脱チェック
Phase 9: PR 作成         → コミット・プッシュ・PR作成
Phase 10: 振り返り       → 空振り・手戻り検出、知見蓄積、改善提案
```

このフローは「公開契約を先に固定し、実装は cluster 単位で解く」ことを重視する。従来の `issue-implement` の I/F先行は維持しつつ、内部 I/F を早期に凍結しない。

## モデル分担

- **Opus**: メインオーケストレータ。Phase 制御、例外判断、Escalation 判定を担当
- **Sonnet**: cluster 実装、failure 分類、採用判定、ゴール整合性、PR作成、振り返り
- **Codex**: 契約レビュー、テスト戦略レビュー、Red レビュー、品質ゲートレビュー
- **Haiku**: コメント整理などの定型変換

## サブエージェント分離原則

- 実装者にレビューをさせない
- レビュアに実装ログ全文を渡さない
- 採用判定者はレビュー結果とコード事実だけを見る
- ゴール整合性判定者は品質レビュー結果に引きずられない

## Phase 0: 初期化

`references/initialization.md` に従い、前提チェック・開発準備・引数解析・計画検出・Issue 読み込みを行う。

## Phase 1: 契約設計

Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/contract-design.md` を読み込み、`references/contract-design.md` に従って公開 I/F、エラー契約、不変条件、生成コード前提、受け入れ条件対応表を定義する。`contract.md` と `contract-decisions.md` を出力する。

契約設計後は `references/contract-review.md` に従い、Codex の独立レビューと Sonnet の採用判定を行う。内部 I/F はこの Phase では凍結しない。

## Phase 2: テスト戦略

`references/test-strategy.md` に従い、変更対象ごとに `UT必須 / 統合テストで十分 / 新規テスト不要` を決める。`docs/guides/testing-strategy.md` を必ず参照する。

戦略策定後は `references/test-strategy-review.md` に従い、Codex の独立レビューと Sonnet の採用判定を行う。

## Phase 3: Red 生成

Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/red-generation.md` を読み込み、`references/red-generation.md` に従って契約からテストまたは検証項目を導出する。全テスト Green を一度に狙うのではなく、後段で failure cluster に分けやすい粒度で生成する。

生成後は `references/red-review.md` に従い、Codex の独立レビューと Sonnet の採用判定を行う。

## Phase 4: 失敗分類

`references/failure-clustering.md` に従い、Red 実行結果を failure cluster に分類する。分類候補は `if_gap`、`compile_gap`、`behavior_bug`、`test_bug`、`tool_behavior_gap`。

原因分類は Sonnet の独立サブエージェントに委譲する。メインエージェントは分類結果を見て cluster の処理順を決める。

## Phase 5: 実装

`references/implementation-loop.md` に従う。実装対象は issue 全体ではなく cluster 単位とする。

- Sonnet サブエージェントで対象 cluster の最小修正を行う
- 検証は狭いテスト → 広いテストの順で進める
- 改善判定は `pass_count` ではなく failure 集合差分で行う
- 改善した cluster だけ checkpoint 化する

## Phase 6: Escalation

`references/escalation.md` に従う。同一 cluster の連続失敗や compile/tool gap の継続時に、契約・テスト・実装のどこに問題があるか再判定する。

修正時は必ず以下を記録する:

- failing evidence
- invariant statement
- minimal patch scope

## Phase 7: 品質ゲート

`references/quality-gate.md` に従う。

- `just test`
- `just fmt`
- `just lint`
- Codex 観点付きレビュー
- Sonnet 採用判定
- simplify
- Haiku によるコメント整理

## Phase 8: ゴール整合性

常に実行する。`references/goal-alignment.md` に従い、受け入れ条件トレーサビリティとスコープ逸脱をチェックする。

## Phase 9: PR 作成

Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/pr-creation.md` を読み込み、コミット・プッシュ・PR作成を自律実行する。

## Phase 10: 振り返り

Phase 9 完了後、または Phase 5〜7 中断後に実行する。Issue 番号を渡して Sonnet モデルのサブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/retrospective.md` を読み込み、`.output/` 内のファイルを自己取得して知見の収集・原因分析・改善提案を行う。
