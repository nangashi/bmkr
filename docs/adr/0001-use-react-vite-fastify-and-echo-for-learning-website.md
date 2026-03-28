---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# ECサイトのフロントエンド・BFF・バックエンドAPIのフレームワーク選定

## Decision Outcome

ECサイトの3層経路（フロントエンド・BFF・バックエンドAPI）に React (Vite) + Fastify + Echo を採用する。要件定義の3層構成に忠実に、BFF を独立したサービスとして物理的に分離することで、API集約・認証ゲートウェイ・レスポンス整形といった BFF の責務を明確に学べる構成とする。Fastify は成熟したプラグインエコシステム（JWT, HTTP Proxy, CORS, Pino ログ等）を持ち、BFF 構築に必要な機能が揃っている。また React + Go は求人需要が高く、Fastify のスキルも実務で活きる。Go のバックエンドフレームワークとして Echo を ECサイト・商品管理・顧客管理の全サービスに適用する。

### Consequences

- React / Fastify / Echo という実務でも通用する技術スタックのスキルが身につく
- BFF が物理的に独立したサービスとなり、ECサイト画面のアクセス経路における責務分離を明確に学べる
- 3つのサーバーを管理する開発環境の構築が必要で、Docker Compose 等の環境整備に初期コストがかかる
- TypeScript と Go 間の型同期は自動化されないため、OpenAPI 等の仕組みを別途導入する必要がある

## Context and Problem Statement

学習用模擬ECサイトを4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービス構成で開発する。このうち ECサイト画面は「ブラウザ → BFF → バックエンドAPI」の3層経路でアクセスされ、BFF は独立したサービスとして定義されている。本 ADR ではこの3層経路（フロントエンド・BFF・バックエンドAPI）のフレームワークを選定する。商品管理・顧客管理の管理画面UIの技術選定は本 ADR のスコープ外とし、別途決定する。

## Prerequisites

- 要件定義（docs/plans/requirements.md）で4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービス構成が定義されている
- ECサイト画面は BFF を経由してバックエンドAPIにアクセスする。商品管理・顧客管理の管理画面は各サービスが直接提供する
- フロントエンド・BFF: TypeScript、バックエンドAPI（ECサイト・商品管理・顧客管理）: Go
- 学習目的のプロジェクトである（商用運用ではない）

## Decision Drivers

- 学習リソースの豊富さ
- 各言語イディオムとの親和性
- アーキテクチャの学習効果
- 実務への応用可能性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **React (Vite) + Fastify + Echo（採用）** | React (Vite) でフロントエンド SPA、Fastify で独立した BFF、Echo でバックエンド API を構築する 3 層分離構成 |
| Next.js + Echo | Next.js (App Router) でフロントエンド + BFF を統合し、Go の Echo でバックエンド API を構築する構成 |
| SvelteKit + Chi | SvelteKit (+server.ts) でフロントエンド + BFF を統合し、Go の Chi でバックエンド API を構築する構成 |
| React (Vite) + Hono + Echo | React (Vite) でフロントエンド SPA、Hono で独立した BFF、Echo でバックエンド API を構築する 3 層分離構成 |

## Comparison Overview

| 観点 | Next.js + Echo | React (Vite) + Fastify + Echo | SvelteKit + Chi | React (Vite) + Hono + Echo |
|------|---------------|-------------------------------|-----------------|---------------------------|
| 学習リソースの豊富さ | ◎ | ○ | △ | △ |
| 各言語イディオムとの親和性 | ○ | ◎ | ○ | ◎ |
| アーキテクチャの学習効果 | △ | ◎ | △ | ◎ |
| 実務への応用可能性 | ◎ | ◎ | △ | ○ |
| BFF 独立性 | △ | ◎ | △ | ◎ |
| 開発環境の複雑さ | ◎ | △ | ◎ | △ |

## Notes

- Next.js は BFF 層が統合されるため、要件定義の3層構成と整合しない。App Router は概念が多く（RSC, Server Actions, Streaming 等）学習コストが高い
- SvelteKit は BFF が統合されており独立サービスとしての運用経験は得られない。Svelte 5 (Runes) により既存チュートリアルの多くが古くなっている
- Hono は急成長中だが求人で名指しされることはまだ少ない。Fastify と同様に TypeScript-Go 間の型同期は自動化されない
- 「React + Fastify + Echo」の統合チュートリアルはほぼ存在せず、各技術を個別に学んで組み合わせる必要がある

## More Information

- 要件定義: docs/plans/requirements.md
- React (Vite) 公式ドキュメント: https://vite.dev/
- Fastify 公式ドキュメント: https://fastify.dev/
- Echo 公式ドキュメント: https://echo.labstack.com/docs
- BFF の責務は `@connectrpc/connect-fastify` による API 集約、`@fastify/jwt` による認証、`@connectrpc/connect-node` によるバックエンドへの通信で実現する（通信プロトコルは ADR-0003 で決定）
- 商品管理・顧客管理の管理画面UIの技術選定は別途 ADR で決定する
