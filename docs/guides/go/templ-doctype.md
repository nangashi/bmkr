---
id: templ-doctype
category: testing
scope: [admin-ui]
severity: low
detectable_by_linter: false
---

# templ の小文字 doctype 出力

## アンチパターン

テストで `<!DOCTYPE html>`（大文字）を期待値として書く。templ は `<!doctype html>`（小文字）を出力するため、テストが失敗する。

## 正しいパターン

templ 生成コードの実際の出力を確認してからテストの期待値を書く。

```go
// Bad
expected := "<!DOCTYPE html><html>"

// Good
expected := "<!doctype html><html>"
```
