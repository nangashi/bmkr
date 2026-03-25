---
globs: services/ec-site/frontend/src/**/*.tsx
---

# Tailwind preflight によるスタイルリセット

## アンチパターン

Tailwind CSS の preflight（CSS リセット）が有効な環境で、`h1`〜`h6`、`a`、`button`、`table` 等のセマンティック HTML 要素をスタイルなしで使う。preflight がブラウザデフォルトスタイルをリセットするため、これらの要素はプレーンテキストと同じ見た目になる。

## 正しいパターン

以下のいずれかで対応する:

1. **`@layer base` で復元**（プロジェクト共通）: `src/index.css` の `@layer base` にデフォルトスタイルを定義する。全コンポーネントに適用される
2. **ユーティリティクラスで明示指定**（個別要素）: `className="text-2xl font-bold"` のように要素ごとに指定する

## 具体例

```css
/* src/index.css — @layer base で復元 */
@layer base {
  h1 {
    font-size: 2em;
    font-weight: bold;
  }
  a {
    color: #2563eb;
    text-decoration-line: underline;
  }
}
```

```tsx
// ユーティリティクラスで明示指定
<button className="rounded border border-border bg-gray-100 px-4 py-1">
  送信
</button>
```

## 新しい要素を追加する前に

`@layer base` で既にスタイルが定義されている要素は以下を確認:
- `src/index.css` の `@layer base` ブロック

未定義の要素（`h2`〜`h6`、`table`、`select` 等）を初めて使う場合は、`@layer base` に追加するか、ユーティリティクラスで明示的にスタイルすること。
