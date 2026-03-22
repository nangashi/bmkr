---
id: cargo-cult-error-handling
category: llm-anti-pattern
scope: [all]
severity: medium
detectable_by_linter: false
---

# 形式的エラーハンドリング

LLM は「エラーハンドリングは多いほど良い」と考え、意味のない処理を追加する傾向がある。

## アンチパターン

**Go:**
- 意味のない `fmt.Errorf("failed to X: %w", err)` の連鎖（文脈情報を追加しない wrap）
- 呼び出し元で必ず処理されるエラーに対する冗長な nil チェック
- 到達しないエラーパスの処理

**TypeScript:**
- catch して re-throw するだけの try-catch
- 意味のない `.catch(err => { throw err })`
- フレームワークが処理するエラーの手動ハンドリング

## 正しいパターン

- エラーを wrap するなら、呼び出し元が判断に使う文脈情報を追加する
- 既存コードのエラーハンドリングパターンを読んでそれに従う
- フレームワーク（Echo, Fastify）のエラーハンドリング機構を信頼する

## エラー情報を握り潰さない

形式的ハンドリングの逆パターンとして、エラーを汎用メッセージに置き換えて元の情報を失うケースもある。

```go
// Bad: 元のエラーが失われ、ログやデバッグで原因追跡できない
return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch products")

// Good: 元のエラーを保持する（Echo がログに記録する）
return echo.NewHTTPError(http.StatusInternalServerError, err)
```

「意味のない wrap を減らす」と「エラー情報を保持する」は両立する。不要な中間層を減らしつつ、最終的なエラーレスポンスでは元の情報を渡す。
