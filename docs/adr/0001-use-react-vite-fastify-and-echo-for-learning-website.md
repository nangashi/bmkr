---
status: "accepted"
date: 2026-03-15
---

# ECサイトのフロントエンド・BFF・バックエンドAPIに React (Vite) + Fastify + Echo を採用する

## Context and Problem Statement

学習用模擬ECサイトを4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービス構成で開発する。このうち ECサイト画面は「ブラウザ → BFF → バックエンドAPI」の3層経路でアクセスされ、BFF は独立したサービスとして定義されている。本 ADR ではこの3層経路（フロントエンド・BFF・バックエンドAPI）のフレームワークを選定する。商品管理・顧客管理の管理画面UIの技術選定は本 ADR のスコープ外とし、別途決定する。

## Prerequisites

* 要件定義（docs/plans/requirements.md）で4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービス構成が定義されている
* ECサイト画面は BFF を経由してバックエンドAPIにアクセスする。商品管理・顧客管理の管理画面は各サービスが直接提供する
* フロントエンド・BFF は TypeScript を使用する
* バックエンドAPI（ECサイト・商品管理・顧客管理）は Go 言語を使用する
* 学習目的のプロジェクトである（商用運用ではない）

## Decision Drivers

* 学習リソースの豊富さ — チュートリアル・公式ドキュメント・コミュニティ記事・動画教材がどれだけ充実しているか
* 各言語イディオムとの親和性 — TypeScript / Go それぞれの慣習的な書き方（型システム、エラーハンドリング等）が自然に身につくか
* アーキテクチャの学習効果 — フロント / BFF / バックエンドの責務分離やレイヤー間の通信パターンを理解しやすいか
* 実務への応用可能性 — 学んだ技術スタックが実際の業務プロジェクトや求人市場でどの程度活きるか

## Considered Options

* Next.js + Echo
* React (Vite) + Fastify + Echo
* SvelteKit + Chi
* React (Vite) + Hono + Echo

## Pros and Cons of the Options

### Next.js + Echo

Next.js (App Router) でフロントエンド + BFF を統合し、Go の Echo でバックエンド API を構築する構成。

* Good, because React + Next.js のエコシステムが圧倒的に大きく、困ったときに情報が見つかりやすい (138K+ Stars)
* Good, because 求人市場での需要が最も高く、学習がそのまま実務スキルに直結する
* Good, because App Router の RSC / Server Actions は現在のフロントエンド開発のトレンドを学べる
* Bad, because BFF 層が Next.js に統合されるため、要件定義の 3 層構成と整合せず、レイヤー間の境界が暗黙的になる
* Bad, because App Router は概念が多く (RSC, Server Actions, Streaming 等)、学習コストが高い
* Bad, because Vercel への依存度が高く、セルフホスト時に一部機能の制約がある

### React (Vite) + Fastify + Echo

React (Vite) でフロントエンド SPA、Fastify で独立した BFF、Echo でバックエンド API を構築する 3 層分離構成。

* Good, because 3 層が物理的に分離され、BFF の責務（API 集約・認証ゲートウェイ・レスポンス整形）を明確に学べる
* Good, because Fastify の Type Provider により JSON Schema が型定義と検証を兼ね、TypeScript の型システムを深く学べる
* Good, because Fastify のプラグインエコシステム（JWT, HTTP Proxy, CORS, Pino ログ等）が成熟しており、BFF 構築に必要な機能が揃っている (35.8K Stars)
* Good, because React + Go という求人需要の高い組み合わせで、Fastify のスキルも実務で活きる
* Bad, because 3 つのサーバーを個別に管理する必要があり、開発環境の構築・運用が複雑
* Bad, because 「React + Fastify + Echo」という組み合わせの統合チュートリアルはほぼ存在しない
* Bad, because TypeScript (BFF) と Go (API) 間の型同期は自動化されず、OpenAPI 等の追加ツールが必要

### SvelteKit + Chi

SvelteKit (+server.ts) でフロントエンド + BFF を統合し、Go の Chi でバックエンド API を構築する構成。

* Good, because Chi は外部依存ゼロ・`net/http` 完全互換で、Go 標準ライブラリの理解が最も深まる (21.8K Stars)
* Good, because `+server.ts` により BFF エンドポイントの境界が Next.js より明示的
* Good, because Svelte 5 のコンパイル時最適化で Web の基礎を理解しやすい
* Bad, because Svelte/SvelteKit の求人市場はまだ限定的で、AI コーディングツールの支援も React より弱い
* Bad, because BFF は SvelteKit に統合されており、独立サービスとしての運用経験は得られない
* Bad, because Svelte 5 (Runes) により既存チュートリアルの多くが古くなっており、学習リソースの質にばらつきがある

### React (Vite) + Hono + Echo

React (Vite) でフロントエンド SPA、Hono で独立した BFF、Echo でバックエンド API を構築する 3 層分離構成。

* Good, because 3 層が物理的に分離され、フロント / BFF / API の責務と通信パターンが最も明確に学べる
* Good, because Hono は TypeScript ファーストで、Zod バリデーションと RPC 型推論により BFF↔フロント間の型安全が自動化される (29.4K Stars)
* Good, because Web Standards 準拠・14KB の超軽量設計で、HTTP の基礎を理解するのに最適
* Good, because マルチランタイム対応 (Node.js/Bun/Deno/Edge) により、モダンな実行環境の違いも学べる
* Bad, because 3 つのサーバーを個別に管理する必要があり、開発環境の構築・運用が最も複雑
* Bad, because Hono は急成長中だが、求人で名指しされることはまだ少ない
* Bad, because TypeScript (BFF) と Go (API) 間の型同期は自動化されず、手動管理または OpenAPI/protobuf が必要

## Decision Outcome

Chosen option: "React (Vite) + Fastify + Echo", because BFF の独立性を重視し、要件定義の3層経路に忠実な構成を実現できるため。また Fastify は成熟したプラグインエコシステムを持ち、業務での利用可能性が高いため。Go のバックエンドフレームワークとして Echo を ECサイト・商品管理・顧客管理の全サービスに適用する。

### Consequences

* Good, because React / Fastify / Echo という実務でも通用する技術スタックのスキルが身につく
* Good, because BFF が物理的に独立したサービスとなり、ECサイト画面のアクセス経路における責務分離を明確に学べる
* Bad, because 3 つのサーバーを管理する開発環境の構築が必要で、Docker Compose 等の環境整備に初期コストがかかる
* Bad, because TypeScript と Go 間の型同期は自動化されないため、OpenAPI 等の仕組みを別途導入する必要がある

## More Information

* 要件定義: docs/plans/requirements.md
* React (Vite) 公式ドキュメント: https://vite.dev/
* Fastify 公式ドキュメント: https://fastify.dev/
* Echo 公式ドキュメント: https://echo.labstack.com/docs
* BFF の責務は Route Handlers による API 集約、`@fastify/jwt` による認証、`@fastify/http-proxy` によるバックエンドへの転送で実現する
* 商品管理・顧客管理の管理画面UIの技術選定は別途 ADR で決定する
