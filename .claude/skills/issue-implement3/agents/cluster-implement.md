# cluster 実装 — サブエージェント詳細手順

failure cluster を1つだけ対象に、最小修正で解消する。issue 全体を一度に解こうとしてはならない。

---

## 前提情報の取得

- `.output/issue-implement3/{issue_number}/contract.md`
- `.output/issue-implement3/{issue_number}/test-strategy.md`
- `.output/issue-implement3/{issue_number}/failure-clusters.md`
- 指定された target cluster
- `docs/guides/implementation-anti-patterns.md`
- 必要なら `docs/guides/go-logging.md`

## 作業ルール

- 公開 API は変更しない
- 変更範囲は target cluster の patch scope に収める
- 明示的に求められていない防御的コードを追加しない
- テストを削除して Green にしない
- 内部 I/F の変更が必要な場合も invariant を壊さない最小変更に留める

## 実装手順

1. target cluster の evidence と hypothesis を読む
2. 直すべき failure だけを特定する
3. 最小修正を実装する
4. 指定された狭い検証を実行する
5. 指定された広い検証を実行する
6. 変更ファイルと結果要約を返す

## 中断条件

以下の場合は自由記述のメモを残して停止する:

- patch scope を超える変更が必要
- 契約やテストの誤りが疑われる
- 同じ failure が繰り返し残る
