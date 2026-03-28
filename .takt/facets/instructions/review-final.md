実装完了後の変更全体を 3 つのレビュー観点でレビューする。

## レビューの進め方

### 1. レビュー観点ファイルの読み込み（必須）

以下の 3 ファイルを Read で読んでください:
- `.claude/skills/issue-implement/references/review-silent-failure.md`
- `.claude/skills/issue-implement/references/review-project-compliance.md`
- `.claude/skills/issue-implement/references/review-security.md`

### 2. 差分の確認

`git diff main` で全変更内容を確認する。

### 3. レビュー実行

各観点の Step 1（Analysis）に従い対象コードの構造を把握し、
Step 2（Findings）に従い指摘を導出する。
必要に応じて、変更対象ファイルの周辺コード、`docs/adr/`、`docs/guides/` も参照する。

### 4. 出力

各 finding について以下を含める:
- finding_id（通し番号）
- perspective（各観点ファイルの「perspective ラベル」を使用）
- severity（CRITICAL / IMPORTANT / LOW）
- file:line
- issue（問題の要約）
- analysis（Step 1 で把握した事実）
- reason（なぜ問題か、対応しない場合の具体的シナリオ）
- suggestion（修正案、コード例付き）

各観点ファイルに「出力時の補足」がある場合はそれにも従う。
指摘が 0 件の場合は「指摘なし」と明記する。
