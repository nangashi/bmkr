# Implementation Guides

特定のシーンで考慮すべきパターン・アンチパターンのカタログ。
LLM が生成するコードで繰り返し発生した問題と、その正しい対処法を記録する。

ファイルパターンに紐づく制約（特定ディレクトリのファイル編集時に常に適用すべきルール）は `.claude/rules/` に配置する。

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
