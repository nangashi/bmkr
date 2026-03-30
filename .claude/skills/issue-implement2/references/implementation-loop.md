# 実装ループ

failure cluster を1つずつ解消する。実装は Sonnet、進捗管理と checkpoint 判断は Opus が行う。

---

## 概要

```
cluster を1つ選ぶ
  → hypothesis を明文化
  → Sonnet で最小修正
  → 狭い検証
  → 広い検証
  → failure 集合差分で改善判定
  → 改善したら checkpoint 化
  → 残る cluster がなくなるまで繰り返す
```

## 開始前のコミット

Phase 4 完了後、実装ループ開始前に成果物をコミットする:

```bash
git add -A && git commit -m "checkpoint: before implementation loop (issue #{number})"
```

## Sonnet 実装エージェントへの指示

Issue 番号と target cluster を渡して Sonnet サブエージェント（`model: sonnet`）を起動する。サブエージェントは `agents/cluster-implement.md` を読み込み、その手順に従って実装する。

親エージェントは以下の情報を渡す:

- `contract.md`
- `test-strategy.md`
- 対象 cluster の evidence / hypothesis / patch scope
- `docs/guides/implementation-anti-patterns.md`
- 必要なら `docs/guides/go-logging.md`
- 狭い検証コマンド
- 広い検証コマンド（通常は `just test`）
- 公開 API は変更しないこと
- 内部 I/F 変更が必要なら invariant を守った最小変更に限定すること
- 明示されていない防御的コードを追加しないこと

## 検証順

1. 狭い検証: 該当テストまたは build 対象のみ
2. 広い検証: `just test`
3. 補助検証: `just fmt`、`just lint`

## 改善判定

`pass_count` は使わない。各 attempt の前後で以下を比較する:

- `resolved_failures`
- `new_failures`
- `remaining_failures`
- `severity_delta`

### 改善とみなす条件

- high severity failure が減る
- new failure がない、または軽微で総体として前進している
- patch scope が invariant を破っていない

### 停滞とみなす条件

- 同じ evidence が残る
- 別 cluster の failure を増やした
- 過剰修正で patch scope が拡大した

## 各 attempt の記録

`.output/issue-implement2/{issue_number}/attempts/attempt-{n}.md` に以下を記録する:

- target cluster
- hypothesis
- changed files
- resolved failures
- new failures
- next action

## cluster 完了条件

- 対象 cluster の failure が消えた
- 他 cluster の failure が増えていない
- 契約・不変条件を維持している
