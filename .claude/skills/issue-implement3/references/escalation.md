# Escalation

同一 cluster の連続失敗時に、実装者バイアスを切って根本原因を再判定する。

---

## 発火条件

- 同一 cluster で2回連続失敗
- `compile_gap` が繰り返される
- `tool_behavior_gap` が繰り返される
- 実装が不自然な複雑さを帯び始めた

## 再判定の観点

- 本当に実装の問題か
- 契約に不足があるか
- テスト期待値やセットアップが誤っているか
- 外部ツールの実挙動を誤解していないか

## 分岐

| 判定 | 対応 |
|------|------|
| `implementation_continue` | Sonnet で最終1回だけ別方針を試す |
| `contract_fix` | Opus が契約を修正し、必要なら契約レビューを再実行する |
| `test_fix` | Opus がテストを修正し、テストレビューを再実行する |
| `stop_and_report` | 現在の best state で停止し、ユーザーに報告する |

## 証拠要件

Escalation で契約またはテストを修正する場合、以下を必ず `.output/issue-implement3/{issue_number}/escalation.md` に記録する:

1. failing evidence
2. invariant statement
3. minimal patch scope

## 注意

- 公開 API の変更は原則禁止
- 公開 API を変える場合は「契約不足」または「ADR違反是正」の場合に限る
- 契約を変えたら契約レビューを再実行する
- テストを変えたらテストレビューを再実行する
