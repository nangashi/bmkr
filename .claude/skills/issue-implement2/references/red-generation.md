# Red 生成

契約とテスト戦略から、テストまたは検証項目を導出する。実装がまだないため、テストは契約追従でなければならない。

---

## 手順

1. `agents/red-review.md` を参照する前に、既存テストパターンを調べる
2. `contract.md` と `test-strategy.md` を読み、対象ごとにテストを書くか検証項目を記録する
3. コンパイルが通ることを確認する
4. `just test` を実行し、失敗結果を取得する
5. `.output/issue-implement2/{issue_number}/red-summary.md` に要約する

## ルール

- 実装に依存した期待値を置かない
- Assertion Roulette を避ける
- 既存テストヘルパーやモックを再利用する
- 新規テスト不要に分類した対象は、検証項目だけを `red-summary.md` に書く

## red-summary.md に含める項目

```markdown
# Red Summary

## Added Tests
- 追加したテスト一覧

## Deferred Checks
- テスト不要または統合テストで担保する項目

## Initial Failures
| 種別 | テスト名またはエラー | 概要 |
|------|----------------------|------|
```
