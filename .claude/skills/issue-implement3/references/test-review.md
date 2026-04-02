# テストレビュー（L3 のみ）

v2 の Red レビューとテスト戦略レビューを 1 回に統合したレビュー。テスト戦略の妥当性とテストコードの品質を同時に検証する。

---

## Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff HEAD`
- `{perspective_files}`: `testability.md`
- `{output_path}`: `.output/issue-implement3/{issue_number}/review-test.md`

### perspectives へのインライン化

testability.md に以下をインラインで追加する（perspectives ファイル自体は変更しない。プロンプト構築時に追記する）:

```
## 追加コンテキスト: テスト戦略サマリー

{test-strategy.md の内容をサマリーとして埋め込み}
```

これにより Codex は 2 つのセクションを同時に評価できる:

1. **テスト戦略の妥当性**: test-strategy.md の分類判断（UT/統合/不要）は適切か、テスト対象の漏れはないか
2. **テストの品質**: テストコードの契約追従・粒度・実装非依存

---

## 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:

- `issue_number`
- `review_output_path`（`.output/issue-implement3/{issue_number}/review-test.md`）
- `output_path`（`.output/issue-implement3/{issue_number}/review-test-filtered.md`）

---

## 判定結果の処理

- 採用指摘あり → Opus がテストまたはテスト戦略を修正する
- 採用指摘なし → Phase 5（失敗分類）に進む
