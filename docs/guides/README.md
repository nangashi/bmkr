# Implementation Guides

実装時に参照すべきパターン・アンチパターンのカタログ。
LLM が生成するコードで繰り返し発生した問題と、その正しい対処法を記録する。

## 構成

```
docs/guides/
├── go/            # Go (Echo, sqlc, templ) 固有のパターン
├── react/         # React/Vite 固有のパターン（今後追加）
├── fastify/       # Fastify/BFF 固有のパターン（今後追加）
└── architecture/  # 技術非依存のアーキテクチャ・LLM行動パターン
```

## 参照タイミング

- **計画時** (issue-plan): architecture/ のパターンを確認し、設計で踏まないか事前チェック
- **テスト作成時** (issue-implement Phase 3): テスト関連パターンの強制学習
- **実装時** (issue-implement): 該当技術のパターンを読んでからコード生成
- **レビュー時** (codex-review, issue-implement Phase 5): パターン違反をチェックリストとして使用

## 運用ルール

- Linter で検出できるものはここに書かない（golangci-lint, oxlint に任せる）
- 新しいパターンを発見したら追記する
- Linter ルール化できたら `detectable_by_linter: true` に更新し、アーカイブを検討する
