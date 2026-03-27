# I/F設計

計画の内容に基づいて実行方法を選択し、型・インターフェース・シグネチャ・動作コメントを定義する。

---

## 実行レベルの判断

計画の内容に基づいて実行方法を選択する。

| 計画の内容 | 実行方法 | 動作コメント |
|-----------|---------|------------|
| 新しい型・インターフェース・関数の定義が必要 | サブエージェント起動（`agents/interface-design.md`） | あり（全ての新規関数に記述） |
| 既存コードの修正で新しいロジック分岐がある | メインが直接編集 | あり（新しい分岐・エッジケースのみ） |
| 機械的な置換・削除のみ | メインが直接編集 | なし |

---

## 動作コメントの `wip:` マーカー

動作コメントには `// wip:` プレフィックスを付ける。このマーカーによりテスト作成フェーズが既存コメントと動作コメントを区別でき、実装フェーズのクリーンアップ対象を特定できる。

**新規関数の場合:**
```go
// wip: HandleProductList handles GET /admin/products.
// wip: 動作:
// wip:   - page パラメータをパース、不正値は 1 にフォールバック
// wip:   - DB エラー時は echo.NewHTTPError(500) を返す
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    panic("not implemented")
}
```

**既存関数内の変更の場合:**
```go
func (h *CartServiceHandler) GetCart(...) {
    // Get or create cart  ← 既存コメント（触らない）

    // wip: 商品取得失敗時に slog.WarnContext でログ出力する（product_id, error）
    slog.WarnContext(ctx, "failed to get product", "product_id", item.ProductID, "error", err)
}
```

---

## ガイド読み込み

`docs/guides/workflow/` 配下で関連するガイドを読む。

---

## サブエージェント起動時

Issue 番号を渡してサブエージェントを起動する。サブエージェントは `agents/interface-design.md` を読み込み、その手順に従って型・インターフェース・関数シグネチャ・動作コメントを定義する。
