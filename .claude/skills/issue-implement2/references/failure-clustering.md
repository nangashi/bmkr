# failure cluster 分類

Red 実行結果を cluster に分け、修正対象を絞り込む。issue 全体を一度に解こうとしてはならない。

---

## 分類候補

- `if_gap`: 契約やシグネチャの不足
- `compile_gap`: import/type/build 不整合
- `behavior_bug`: 実装で解消すべきロジック不備
- `test_bug`: テスト前提や期待値の誤り
- `tool_behavior_gap`: sqlc/templ/buf など外部ツール挙動との差

## 手順

1. `just test` の失敗を一覧化する
2. 同じ根本原因で説明できる失敗をまとめる
3. 各 cluster について以下を記録する:
   - cluster id
   - failure type
   - failing evidence
   - root cause hypothesis
   - candidate patch scope
4. Sonnet の独立サブエージェントに分類レビューをさせる
5. `.output/issue-implement2/{issue_number}/failure-clusters.md` を確定する

## 出力フォーマット

```markdown
# Failure Clusters

## Cluster C1
- type: behavior_bug
- evidence:
- hypothesis:
- patch_scope:
- priority:
```

## 優先順

以下の順で処理する:

1. `compile_gap`
2. `if_gap`
3. `behavior_bug`
4. `tool_behavior_gap`
5. `test_bug`
