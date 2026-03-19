---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# 商品管理・顧客管理の管理画面UIに templ + HTMX を採用する

## Context and Problem Statement

学習用模擬ECサイトの商品管理・顧客管理サービスにおいて、管理画面のフロントエンド技術を選定する必要がある。ECサイト画面は React (Vite) + Fastify BFF 構成で決定済み（ADR-0001）だが、管理画面は各 Go (Echo) サービスが直接提供するため、異なる技術選定が求められる。この ADR では管理画面のフレームワーク・ビルドツール・Go からの配信方式を決定する。デザインシステム・CSS フレームワークは扱わない。

## Prerequisites

* ECサイト画面は React (Vite) + Fastify BFF 構成を採用済み (ADR-0001)
* バックエンド API は Go (Echo) を使用する (ADR-0001)
* サービス間通信は Connect RPC (Protobuf) を使用する (ADR-0003)
* 管理画面の認証は echo-jwt を使用する (ADR-0005)
* 管理画面は各 Go サービスが直接提供する（BFF を経由しない）
* 学習目的のプロジェクトである

## Decision Drivers

* Go エコシステムとの統合の自然さ — Echo ミドルウェア・ハンドラとの親和性、単一バイナリ配信の容易さ、echo-jwt との自然な連携
* ECサイト (React) との技術差別化による学習幅 — 異なるパラダイム（SSR/ハイパーメディア vs SPA）を学ぶことで Web 開発全体の理解が深まるか
* テンプレート/UIの型安全性と開発体験 — コンパイル時エラー検出、IDE 補完、リファクタリングの容易さ
* 管理画面ユースケースへの適合性 — テーブル表示・フォーム送信・CRUD 操作など管理画面に典型的な操作の実装しやすさ

## Considered Options

* templ + HTMX
* templ + Datastar
* html/template + HTMX
* React (Vite) SPA を Echo で静的配信

### Excluded Options

* GoAdmin / Go Advanced Admin — 自動生成型は学習目的に合わない。Connect RPC との統合が困難
* gomponents — templ と比較してコミュニティが小さく (1.8K Stars)、HTML ツール連携が弱い

## Comparison Overview

| 判断軸 | templ + HTMX | templ + Datastar | html/template + HTMX | React (Vite) SPA |
|--------|:---:|:---:|:---:|:---:|
| Go エコシステム統合 | ◎ Echo 公式例あり・Go コンパイル・単一バイナリ | ○ templ は良好だが Datastar の Echo 統合例が少ない | ◎ Go 標準ライブラリ・Echo 公式サポート・ビルドツール不要 | △ go:embed で配信可能だが Node.js ビルド依存 |
| 技術差別化・学習幅 | ◎ SSR + ハイパーメディアで React とは正反対のパラダイム | ◎ SSE 駆動は React/HTMX 双方と異なる第3のアプローチ | ◎ SSR + ハイパーメディアで React と正反対 | △ ECサイトと同一技術で差別化が弱い |
| 型安全性・開発体験 | ○ templ はコンパイル時型検査・LSP あり。HTMX 属性は文字列 | ○ templ 型安全 + ReadSignals 型変換。data-* 属性は文字列 | △ ランタイムエラーのリスク・IDE 補完弱い | ◎ TypeScript + React で最高の型安全性・IDE サポート |
| 管理画面適合性 | ○ CRUD/テーブル/フォームは HTMX の得意領域。複雑UIは Alpine.js 追加 | △ テーブル/データグリッドが未成熟。DatastarUI は 15 コンポーネントのみ | ○ CRUD/テーブルは問題なし。リッチ UI コンポーネントは少ない | ○ UI ライブラリ豊富だが管理画面には過剰 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### templ + HTMX

Go 向け型安全テンプレートエンジン templ (10.2K Stars) と HTMX (47.6K Stars) の組み合わせ。GoTH スタック (Go/Templ/HTMX) として確立されたコミュニティがある。

* Good, because templ は Go コードにコンパイルされ、Echo ハンドラから直接 templ コンポーネントを返せる。echo-jwt ミドルウェアとも Cookie ベースで自然に連携し、`go:embed` で単一バイナリ配信が可能
* Good, because React (SPA/クライアント状態管理) とは正反対の SSR + ハイパーメディア駆動を学べ、Web 開発の2大パラダイムを同一プロジェクトで体験できる
* Good, because templ はコンパイル時型チェック + LSP/VSCode 拡張があり、Go ツールチェーンでリファクタリング可能
* Good, because CRUD・テーブル・フォーム・ページネーションは HTMX の得意領域で、JavaScript をほぼ書かずに実装可能
* Good, because GoTH スタックとして確立されたコミュニティがあり、チュートリアル・実運用レポートが充実
* Bad, because HTMX の `hx-*` 属性は文字列ベースで型安全ではなく、LLM/AI コーディング支援は React エコシステムより弱い
* Bad, because 日付ピッカー・リッチテキストエディタなど複雑な UI には Alpine.js 等の追加が必要になる場合がある
* Bad, because `templ generate` のビルドステップが必要で、ホットリロード環境のセットアップが Node.js より複雑

### templ + Datastar

templ に新興のハイパーメディアフレームワーク Datastar (4.2K Stars) を組み合わせる構成。HTMX + Alpine.js の機能を 11KB の単一ライブラリに統合し、SSE ベースでリアルタイム更新に強い。

* Good, because SSE ベースのサーバー駆動型 UI は React (SPA) / HTMX (AJAX) いずれとも異なる第3のアプローチで、学習の幅が最も広い
* Good, because HTMX + Alpine.js の機能を 11KB の単一ライブラリに統合しており、追加ライブラリ不要でリアクティブ UI が実現できる
* Good, because templ の型安全性 + Datastar の `ReadSignals` による Go 構造体デシリアライズで、フロント-バック間のデータ型が保証される
* Bad, because v1.0.0-RC.8 でまだ安定版ではなく、破壊的変更のリスクがあり、トラブルシュート情報が限定的
* Bad, because Echo 固有の統合例が公式にはなく、SSE と Echo ミドルウェアの相互作用は自前検証が必要
* Bad, because テーブル/データグリッドコンポーネントが未提供で、管理画面の中核機能であるページネーション・ソート・フィルタリングの確立パターンがない
* Bad, because SSE 長時間接続中の JWT 期限切れへの対応設計が別途必要

### html/template + HTMX

Go 標準ライブラリのテンプレートエンジンと HTMX の組み合わせ。外部依存ゼロで Go の標準的なやり方を学べる。

* Good, because Go 標準ライブラリのみで外部依存ゼロ。Echo の `echo.Renderer` インターフェースで公式サポートされ、フロントエンドビルドツールが一切不要
* Good, because React とは正反対の SSR + ハイパーメディアパラダイムで、「Go の標準的なやり方」でのテンプレートレンダリングを深く学べる
* Good, because `go:embed` + `template.ParseFS()` で単一バイナリ配信が可能。将来的に templ への段階的移行パスがある
* Bad, because テンプレート内部に型安全性がなく、`{{.DoesNotExist}}` がランタイムエラーになる。IDE 補完・リファクタリング支援も弱い
* Bad, because gopls のテンプレートサポートは実験的で、LSP による支援は templ に大きく劣る
* Bad, because 管理画面に必要な UI コンポーネントエコシステムは React に比べ圧倒的に少ない

### React (Vite) SPA を Echo で静的配信

ECサイトと同じ React (Vite) でビルドした SPA を `go:embed` + `echo.StaticFS()` で Go バイナリに組み込み配信する構成。

* Good, because TypeScript + React エコシステムにより最高レベルの型安全性・IDE サポート・LLM/AI コーディング支援が得られる
* Good, because `@connectrpc/connect-web` + `@connectrpc/connect-query` で TanStack Query と統合した型安全な RPC 呼び出しが可能
* Good, because SPA と API が同一オリジンで配信されるため CORS 設定不要。React の豊富な UI ライブラリが使える
* Bad, because ECサイトと同じ React + Vite であり、SSR やハイパーメディアといった異なるパラダイムを学ぶ機会を完全に逸する
* Bad, because ビルドに Node.js が必要で、Go だけでは閉じない。「Go サービスが直接提供する」という前提の意義が薄れる
* Bad, because BFF がないため SPA 側で JWT の取得・保存・リフレッシュを自前実装する必要があり、echo-jwt との連携が他選択肢より複雑

## Decision Outcome

Chosen option: "templ + HTMX", because Go エコシステムとの統合の自然さと ECサイト (React) との技術差別化による学習幅のバランスを重視したため。

### Rationale

Go エコシステムとの統合の自然さと技術差別化による学習幅の2軸を最も重視した。templ は Go コードにコンパイルされるため Echo のミドルウェアチェーン・echo-jwt・`go:embed` とネイティブに統合でき、「Go サービスが管理画面を直接提供する」という前提に最も忠実な構成を実現できる。また ECサイトの React (SPA) とは正反対の SSR + ハイパーメディアパラダイムを採用することで、Web 開発の2大アプローチを同一プロジェクトで体験でき、学習効果が最大化される。GoTH スタックとして成熟したコミュニティがあり、全前提条件との互換性も完全である。

### Accepted Tradeoffs

なし

### Consequences

* Good, because templ は Go コードにコンパイルされ Echo とネイティブに統合でき、echo-jwt / `go:embed` との連携も自然で、Go サービスが管理画面を直接提供する前提に最も忠実な構成が実現できる
* Good, because ECサイトの React (SPA) とは正反対の SSR + ハイパーメディア駆動を学べ、Web 開発の2大パラダイムを同一プロジェクトで体験できる
* Bad, because HTMX の `hx-*` 属性は文字列ベースで型安全ではなく、日付ピッカー等の複雑な UI には Alpine.js 等の追加が必要になる場合がある

## More Information

* templ 公式ドキュメント: https://templ.guide/
* HTMX 公式サイト: https://htmx.org/
* templ + Echo 統合ガイド: https://templ.guide/integrations/web-frameworks/
* go-echo-templ-htmx サンプル: https://github.com/emarifer/go-echo-templ-htmx
* templ は Go コードにコンパイルされるため、Echo ハンドラ内で直接 templ コンポーネントを呼び出し、`go:embed` で静的アセットを含めた単一バイナリとして配信する
