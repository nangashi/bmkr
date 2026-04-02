---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# 管理画面 UI フレームワークの選定

## Decision Outcome

templ + HTMX を採用する。

Go エコシステムとの統合の自然さと技術差別化による学習幅の2軸を最も重視した。templ は Go コードにコンパイルされるため Echo のミドルウェアチェーン・echo-jwt・`go:embed` とネイティブに統合でき、「Go サービスが管理画面を直接提供する」という前提に最も忠実な構成を実現できる。ECサイトの React (SPA) とは正反対の SSR + ハイパーメディアパラダイムを採用することで、Web 開発の2大アプローチを同一プロジェクトで体験でき、学習効果が最大化される。GoTH スタックとして成熟したコミュニティがあり、全前提条件との互換性も完全である。

### Accepted Tradeoffs

なし

### Consequences

- templ は Go コードにコンパイルされ Echo とネイティブに統合でき、echo-jwt / `go:embed` との連携も自然で、Go サービスが管理画面を直接提供する前提に最も忠実な構成が実現できる
- ECサイトの React (SPA) とは正反対の SSR + ハイパーメディア駆動を学べ、Web 開発の2大パラダイムを同一プロジェクトで体験できる
- HTMX の `hx-*` 属性は文字列ベースで型安全ではなく、日付ピッカー等の複雑な UI には Alpine.js 等の追加が必要になる場合がある

## Context and Problem Statement

商品管理・顧客管理サービスにおいて、管理画面のフロントエンド技術を選定する必要がある。ECサイト画面は React (Vite) + Fastify BFF 構成で決定済み（ADR-0001）だが、管理画面は各 Go (Echo) サービスが直接提供するため、異なる技術選定が求められる。フレームワーク・ビルドツール・Go からの配信方式を決定する。デザインシステム・CSS フレームワークは扱わない。

## Prerequisites

- ECサイト画面: React (Vite) + Fastify BFF (ADR-0001)、バックエンド: Go (Echo) (ADR-0001)
- サービス間通信: Connect RPC (ADR-0003)、管理画面認証: echo-jwt (ADR-0005)
- 管理画面は各 Go サービスが直接提供（BFF を経由しない）、学習目的

## Decision Drivers

- Go エコシステムとの統合の自然さ
- ECサイト (React) との技術差別化による学習幅
- テンプレート/UIの型安全性と開発体験
- 管理画面ユースケースへの適合性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **templ + HTMX（採用）** | Go 向け型安全テンプレート + ハイパーメディア駆動の SSR 構成 |
| templ + Datastar | templ + SSE ベースの新興ハイパーメディアフレームワーク |
| html/template + HTMX | Go 標準ライブラリテンプレート + HTMX |
| React (Vite) SPA | ECサイトと同じ React SPA を Echo で静的配信 |

除外: GoAdmin/Go Advanced Admin（自動生成型、Connect RPC 統合困難）、gomponents（コミュニティが小さく HTML ツール連携が弱い）

## Comparison Overview

| 判断軸 | templ + HTMX | templ + Datastar | html/template + HTMX | React (Vite) SPA |
|--------|:---:|:---:|:---:|:---:|
| Go エコシステム統合 | ◎ Echo 公式例あり・Go コンパイル・単一バイナリ | ○ templ は良好だが Datastar の Echo 統合例が少ない | ◎ Go 標準ライブラリ・Echo 公式サポート・ビルドツール不要 | △ go:embed で配信可能だが Node.js ビルド依存 |
| 技術差別化・学習幅 | ◎ SSR + ハイパーメディアで React とは正反対のパラダイム | ◎ SSE 駆動は React/HTMX 双方と異なる第3のアプローチ | ◎ SSR + ハイパーメディアで React と正反対 | △ ECサイトと同一技術で差別化が弱い |
| 型安全性・開発体験 | ○ templ はコンパイル時型検査・LSP あり。HTMX 属性は文字列 | ○ templ 型安全 + ReadSignals 型変換。data-* 属性は文字列 | △ ランタイムエラーのリスク・IDE 補完弱い | ◎ TypeScript + React で最高の型安全性・IDE サポート |
| 管理画面適合性 | ○ CRUD/テーブル/フォームは HTMX の得意領域。複雑UIは Alpine.js 追加 | △ テーブル/データグリッドが未成熟。DatastarUI は 15 コンポーネントのみ | ○ CRUD/テーブルは問題なし。リッチ UI コンポーネントは少ない | ○ UI ライブラリ豊富だが管理画面には過剰 |

## Notes

- **Datastar は安定版未到達**: v1.0.0-RC.8 でまだ安定版ではなく、破壊的変更のリスクがあり、テーブル/データグリッドコンポーネントが未提供
- **Datastar の SSE + JWT 期限切れ問題**: SSE 長時間接続中の JWT 期限切れへの対応設計が別途必要
- **html/template の型安全性の欠如**: `{{.DoesNotExist}}` がランタイムエラーになり、gopls のテンプレートサポートは実験的
- **React SPA は Go 完結性を損なう**: Node.js ビルドが必要で「Go サービスが直接提供する」前提の意義が薄れる

## More Information

- templ 公式ドキュメント: https://templ.guide/
- HTMX 公式サイト: https://htmx.org/
- templ + Echo 統合ガイド: https://templ.guide/integrations/web-frameworks/
- go-echo-templ-htmx サンプル: https://github.com/emarifer/go-echo-templ-htmx
- templ は Go コードにコンパイルされるため、Echo ハンドラ内で直接 templ コンポーネントを呼び出し、`go:embed` で静的アセットを含めた単一バイナリとして配信する
