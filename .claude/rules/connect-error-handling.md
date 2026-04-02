---
globs: ["services/**/handler*.go"]
---

# Connect RPC のエラーハンドリング

## アンチパターン1: 内部エラーの露出

`connect.NewError(connect.CodeInternal, err)` で DB エラーや外部ライブラリのエラーをそのままクライアントに返す。テーブル名やクエリ構造、ライブラリのバージョン情報などが漏洩する。

## アンチパターン2: エラーコードの不正確な分類

複数の原因でエラーを返しうる関数の結果を、すべて同じエラーコードで返す。たとえば `bcrypt.GenerateFromPassword` は入力超過（72 バイト超）でも内部エラー（メモリ不足等）でも error を返すが、すべてを `CodeInvalidArgument` で返すと、サーバー内部の障害が「クライアントの入力が不正」として扱われる。

## 正しいパターン

1. `CodeInternal` で返すエラーは、生の error を渡さず汎用メッセージを使う
2. エラーの原因が複数ありえる場合は、原因を判定してコードを分ける

## 具体例

```go
// Bad: 生エラーをクライアントに返す
return nil, connect.NewError(connect.CodeInternal, err)

// Good: サーバーログに記録し、クライアントには汎用メッセージ
slog.ErrorContext(ctx, "db query failed", "error", err)
return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
```

```go
// Bad: 全エラーを同じコードで返す
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return nil, connect.NewError(connect.CodeInvalidArgument, err)
}

// Good: 原因を判定してコードを分ける
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    if errors.Is(err, bcrypt.ErrPasswordTooLong) {
        return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("password too long"))
    }
    slog.ErrorContext(ctx, "bcrypt failed", "error", err)
    return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
}
```
