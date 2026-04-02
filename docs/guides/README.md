# Implementation Guides

特定のシーンで考慮すべきパターン・アンチパターンのカタログ。
LLM が生成するコードで繰り返し発生した問題と、その正しい対処法を記録する。

## `.claude/rules/` との棲み分け

| | `.claude/rules/` | `docs/guides/` |
|---|---|---|
| **性質** | 硬い制約（違反したら即リジェクト） | 判断の指針（迷ったときのヒューリスティクス） |
| **ロード** | 常にコンテキストに注入される | スキルが特定フェーズで選択的に参照 |
| **書き方** | 簡潔に（コンテキストコストを意識） | 理由・具体例を詳述してよい |
| **判断基準** | 機械的に適合/違反を判定できる | 状況に応じた判断が必要 |

**迷ったときの問い**: 「これに違反したコードを見たら、文脈を問わず即座にリジェクトするか？」→ Yes なら rules、No なら guides。

## 構成

```
docs/guides/
├── review/          # レビュー・計画時に参照するアンチパターン
├── workflow/        # 作業手順・ワークフロー
└── implementation/  # 実装時のリファレンス
```

## 参照タイミング

- **計画レビュー時** (issue-plan): review/ 配下で関連するガイドを参照
- **設計レビュー時** (issue-implement Phase 2): review/ 配下で関連するガイドを参照
- **I/F設計時** (issue-implement Phase 1): workflow/ 配下で関連するガイドを参照
- **実装時** (issue-implement Phase 4): implementation/ 配下で関連するガイドを参照
- **最終レビュー時** (issue-implement Phase 4 最終ゲート): review/ 配下で関連するガイドを参照

## 運用ルール

- Linter で検出できるものはここに書かない（golangci-lint, oxlint に任せる）
- 新しいパターンを発見したら追記する
- ファイルパターンに紐づく制約は `.claude/rules/` に追加する
- Linter ルール化できたら `detectable_by_linter: true` に更新し、アーカイブを検討する
