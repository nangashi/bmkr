---
globs: ["**/db/queries/*.sql"]
---

# 変更系クエリは :execrows を使う

## アンチパターン

UPDATE / DELETE クエリを `:exec` で定義する。影響行数が 0 でもエラーにならず、「更新/削除したつもりが実際には何も変わっていない」サイレントフェイルが発生する。

事前に SELECT で存在確認しても、確認と変更の間に別リクエストが割り込む（TOCTOU ギャップ）ため、存在確認だけでは防げない。

## 正しいパターン

変更系クエリは `:execrows` で定義し、ハンドラ側で影響行数を検査する。

## 具体例

```sql
-- Bad: 影響行数を返さない
-- name: UpdateCartItemQuantity :exec
UPDATE cart_items SET quantity = $1 WHERE id = $2 AND cart_id = $3;

-- Good: 影響行数を返す
-- name: UpdateCartItemQuantity :execrows
UPDATE cart_items SET quantity = $1 WHERE id = $2 AND cart_id = $3;
```

```go
// Bad: 0 件更新でも err == nil → サイレントフェイル
err := h.q.UpdateCartItemQuantity(ctx, params)

// Good: 0 件を検出して NOT_FOUND を返す
rows, err := h.q.UpdateCartItemQuantity(ctx, params)
if err != nil {
    return nil, connect.NewError(connect.CodeInternal, ...)
}
if rows == 0 {
    return nil, connect.NewError(connect.CodeNotFound, errors.New("item not found"))
}
```

## 例外

- 冪等な削除（「なければ何もしない」が正しい動作）は `:exec` でよい
- RETURNING 句で結果を返す場合は `:one` / `:many` を使う
