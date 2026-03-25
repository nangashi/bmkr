トリアージで accept と判定された指摘を修正する。

## 入力

`{report:triage-results.md}` から accept 指摘を読む。
オシレーション回避指示がある場合: `{report:oscillation-directives.md}` を Read し、
記載された directive に従う（過去に打ち消しが検出された変更を再度行わない）。

## 修正手順

1. accept 指摘のみを対象に修正する（reject は対応しない）
2. 各指摘について、指摘されたファイル・行を Read で確認してから修正する
3. 修正後 `just test` + `just fmt` + `just lint` を実行して問題がないことを確認する
4. テストが失敗した場合は、修正が既存の動作を壊していないか確認し調整する
