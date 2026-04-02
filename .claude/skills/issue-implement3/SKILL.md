---
name: issue-implement3
description: issue-plan の計画をもとに、変更規模に応じた適応的フローで実装・テスト・レビュー・PR作成まで行う。`/issue-implement3` で起動。
allowed-tools: Bash, Read, Glob, Grep, Agent, Write, Edit, AskUserQuestion, Skill
disable-model-invocation: true
argument-description: "Issue番号。例: '#123', '123'"
---

# issue-implement3

GitHub Issue の実装計画をもとに、変更規模に応じたフローで実装から PR 作成まで自動で進める。

## 前提

- `/issue-plan` で計画が作成済みであること
- 計画は Issue コメントに `<!-- issue-plan:plan:done -->` マーカーで記録されている
- 実装不可の問題がなければ、PR 作成まで対話なしで自動実行する

## 設計方針

1. **変更規模に応じた適応的フロー**: Phase 0 で変更をレベル分類し、対応するフローファイルを読み込む。小規模変更に不要な Phase を実行しない
2. **progressive disclosure**: SKILL.md は分類ロジックのみ。各レベルのフロー詳細は独立ファイルに分離し、実行時に必要なファイルだけを読む
3. **v1 の効率性 + v2 の品質保証**: L1/L2 は v1 の軽量フロー、L3 は v2 の contract.md・failure clustering・Escalation を活用

## Phase 0: 初期化 + レベル分類

`references/initialization.md` に従い、前提チェック・開発準備・引数解析・計画検出・Issue 読み込みを行う。

### レベル分類

計画コメントの内容から以下を **順に** 判定する:

1. 既存の公開 I/F を変更するか? → Yes なら **L3**
2. 不変条件の明文化が必要か?（残す I/F、変更しないレスポンス形状） → Yes なら **L3**
3. DB スキーマ変更を含むか? → Yes なら **L3**
4. 影響範囲が広いか? → Yes なら **L3 を検討**。以下のコマンドで確認する:
   - `git diff --stat main` で変更予定ファイル数を確認（計画のステップから推定でもよい）。**5 ファイル以上**なら L3 候補
   - 変更する関数名で `grep -rn '{関数名}' --include='*.go'` を実行し呼び出し元を確認。**3 箇所以上**から呼ばれていれば L3 候補
   - いずれかに該当し、かつ変更が波及する可能性がある場合は L3 とする
5. 新規の型・インターフェース・関数シグネチャの定義が必要か? → Yes なら **L2**
6. 上記いずれも No → **L1**

※ 閾値（5 ファイル / 3 呼び出し元）はヒューリスティックであり、Phase 8（振り返り）で L2→L3 昇格頻度を追跡して調整する。

1 Issue に L2 と L3 の変更が混在する場合は **L3** として扱う。推奨は issue-plan 段階での分割。

分類結果を `scope.md` に記録する:

```markdown
## 変更レベル: L2（局所的）
### 根拠
- 新規 RPC AddItem の追加（新規 I/F 定義あり）
- 既存 I/F への変更なし
- DB スキーマ変更なし
### 実行パス
references/flow-L2.md
```

### フローファイルの読み込み

分類結果に応じて **1 つだけ** 読み込み、以降はそのファイルの指示に従う:

| レベル | 読み込むファイル |
|--------|------------------|
| L1: 機械的 | `references/flow-L1.md` |
| L2: 局所的 | `references/flow-L2.md` |
| L3: 構造的 | `references/flow-L3.md` |

## モデル分担

| 役割 | モデル | 理由 |
|------|--------|------|
| メインオーケストレータ | Opus | Phase 制御、レベル分類、例外判断 |
| 設計・実装・テスト導出 | Sonnet | bounded な入力からの生成タスク |
| 独立レビュー | Codex | 実装者と別の目。perspectives は自己完結型 |
| 採用判定 | Sonnet | レビュー結果とコード事実のみを見る |
| 設計/契約の修正 | Opus | 判断を伴う修正 |
| 品質ゲートの修正 | Sonnet | 指摘に沿った bounded な修正 |
| コメント整理 | Haiku | 定型変換 |
| ゴール整合性・PR・振り返り | Sonnet | 構造化された照合・整形タスク |

## サブエージェント分離原則

- 実装者にレビューをさせない
- レビュアに実装ログ全文を渡さない
- 採用判定者はレビュー結果とコード事実だけを見る
- ゴール整合性判定者は品質レビュー結果に引きずられない

## 実装ルール（全レベル共通）

`references/implementation-rules.md` に定義。以下は特に重要なもの:

- テスト期待値を実装に合わせて変更することは禁止
- 既存テストを削除して Green にすることは禁止
- 例外は仕様廃止・DI 方式変更に伴う再設計のみ（invariant statement を記録）
