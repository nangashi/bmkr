# 設計レビュー（Phase 2）

I/F定義のみを対象にレビューする。Codex（汎用）とプロジェクト適合性レビュー（Sonnet サブエージェント）を **並列実行** する。

### 2a. Codex 汎用レビュー（Codex CLI）

メインエージェントが以下を実行する:

```bash
codex exec review --uncommitted -o .output/issue-implement/{issue_number}/review-design-codex.md
```

### 2b. プロジェクト適合性レビュー（Sonnet サブエージェント）

このセクションはサブエージェントへの指示として使う。サブエージェントはこのセクションを読み込んで作業する。

**やること:**

1. `git diff HEAD` で Phase 1 の変更内容を確認する
2. 変更対象ファイルの周辺にある既存コードを Read して、既存パターンを把握する
3. `docs/adr/` 配下の ADR を読み、決定事項を把握する
4. `docs/guides/` 配下のガイドを全て読み、実装パターンを把握する
5. 以下の観点でレビューする:
   - **命名規則**: 関数名・型名・変数名が周辺の既存コードと一致しているか
   - **ファイル配置**: handler, types, templates 等の分け方が同一サービス内の既存構造に揃っているか
   - **ガイド準拠**: `docs/guides/` のパターンに違反していないか
   - **ADR 準拠**: `docs/adr/` の決定事項に反していないか
6. 結果を `.output/issue-implement/{issue_number}/review-design-project.md` に Write で書き出す
