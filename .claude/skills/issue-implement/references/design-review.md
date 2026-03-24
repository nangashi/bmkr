# 設計レビュー（Phase 2）

I/F定義のみを対象に、Codex CLI で 3 観点（型設計・エラー契約・テスト駆動性）のレビューを実施する。

### 2a. Codex 観点付きレビュー（Codex CLI）

メインエージェントが以下のプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

```
以下の git diff の変更内容をコードレビューしてください。

差分:
git diff HEAD

## レビューの進め方
1. まず以下のレビュー観点ファイルを読んでください:
   - .claude/skills/issue-implement/references/review-type-design.md
   - .claude/skills/issue-implement/references/review-error-contract.md
   - .claude/skills/issue-implement/references/review-testability.md
2. 各観点の Step 1（Analysis）に従い、対象コードの構造を把握してください
3. 各観点の Step 2（Findings）に従い、指摘を導出してください
4. 必要に応じて、変更対象ファイルの周辺コード、docs/adr/、docs/guides/review/ も参照してください

## 出力
結果を .output/issue-implement/{issue_number}/review-design.md に書き出してください。

### フォーマット
観点ごとにセクションを分け、各 finding について以下を含めてください:
- finding_id（通し番号）
- perspective（各観点ファイルの「perspective ラベル」を使用）
- severity（high / medium / low）
- file:line
- issue（問題の要約）
- analysis（Step 1 で把握した事実）
- reason（なぜ問題か、対応しない場合の具体的シナリオ）
- suggestion（修正案、コード例付き）

各観点ファイルに「出力時の補足」がある場合はそれにも従ってください。
```

プロンプト中の `{issue_number}` は実際の Issue 番号に置換すること。`git diff HEAD` の出力はプロンプト構築時に展開して埋め込む。
