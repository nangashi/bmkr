# Codex レビュープロンプトテンプレート

契約レビュー、テスト戦略レビュー、Red レビュー、品質ゲートレビューで共通のプロンプト構造。呼び出し元がパラメータを埋めて `timeout 300 codex exec --full-auto` に stdin で渡す。

---

## パラメータ

| パラメータ | 例 |
|-----------|----|
| `{diff_command}` | `git diff HEAD` / `git diff main` |
| `{perspective_files}` | `type-design.md`, `testability.md` など |
| `{output_path}` | `.output/issue-implement2/{issue_number}/review-*.md` |

## テンプレート

```
以下の git diff の変更内容をコードレビューしてください。

差分:
{diff_command}

## レビューの進め方
1. まず以下のレビュー観点ファイルを読んでください:
{perspective_files}
2. 各観点の Step 1（Analysis）に従い、対象コードの構造を把握してください
3. 各観点の Step 2（Findings）に従い、指摘を導出してください
4. 必要に応じて、変更対象ファイルの周辺コード、docs/adr/、docs/guides/ も参照してください

## 出力
結果を {output_path} に書き出してください。

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
```

## 構築時の注意

- `{diff_command}` の出力はプロンプト構築時に展開して埋め込む
- `{issue_number}` は実際の Issue 番号に置換する
- `{perspective_files}` は `   - .claude/skills/issue-implement2/perspectives/{ファイル名}` の形式で列挙する
