# 振り返り — サブエージェント詳細手順

プロセスの過剰・不足を検出し、失敗・成功から得た知見を収集・分析して改善提案を作成する。

---

## 前提情報の自己取得

以下のファイルを `.output/issue-implement2/{issue_number}/` から Read する（存在しないものはスキップ）:

| ファイル | 内容 |
|---------|------|
| `issue-context.md` | Issue の前提情報 |
| `scope.md` | スコープ定義 |
| `contract.md` | 公開契約 |
| `contract-decisions.md` | 計画との判断差分 |
| `review-contract.md` / `review-contract-filtered.md` | 契約レビュー結果と採用判定 |
| `test-strategy.md` | テスト戦略 |
| `review-test-strategy.md` / `review-test-strategy-filtered.md` | テスト戦略レビュー結果と採用判定 |
| `red-summary.md` | Red の結果 |
| `review-red.md` / `review-red-filtered.md` | Red レビュー結果と採用判定 |
| `failure-clusters.md` | failure 分類 |
| `attempts/` 配下 | 実装ループの試行ログ |
| `escalation.md` | Escalation の証拠と判定 |
| `review-final.md` / `review-final-filtered.md` | 品質ゲートレビュー結果と採用判定 |
| `oscillation-directives.md` | 振動検出の directive |
| `goal-alignment.md` | ゴール整合性チェック結果 |

## 分析

- どの cluster で手戻りが発生したか
- 誤診はどこで起きたか
- 契約不足、テスト不足、実装不足のどれが多かったか
- 現行フローで不要に重かった Phase はどこか

## 改善提案

以下のいずれかを提案する:

- `docs/guides/` への追記
- `.claude/rules/` へのルール追加
- スキル改善
- issue-plan へのフィードバック

各提案に期待効果とコストを1行で添える。
