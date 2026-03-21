---
id: state-mutation-refetch
category: logic
scope: [all]
severity: high
detectable_by_linter: false
---

# 状態変更後のデータ再取得

## アンチパターン

状態を変更（クランプ、正規化、更新等）した後、変更前のデータをそのまま使う。典型例: ページ番号をクランプしたが、クランプ前のページ番号で取得したデータを返す。

## 正しいパターン

状態を変更したら、変更後の値でデータを再取得する。

## 具体例

```go
// Bad: クランプ後に再フェッチしない
products, _ := h.q.ListProducts(ctx, db.ListProductsParams{Offset: offset(page)})
if page > totalPages {
    page = totalPages // クランプしたが products は古い page で取得済み
}

// Good: クランプ後に再フェッチ
if page > totalPages {
    page = totalPages
    products, _ = h.q.ListProducts(ctx, db.ListProductsParams{Offset: offset(page)})
}
```
