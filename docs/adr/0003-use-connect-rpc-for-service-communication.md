---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# マイクロサービス間の通信プロトコルに Connect RPC を採用する

## Context and Problem Statement

React (Vite) + Fastify BFF + Go Echo バックエンド API の 3 層構成において、フロント↔BFF 間および BFF↔バックエンド間の通信プロトコルを選定する必要がある。2つの境界で異なるプロトコルを採用する可能性も含めて検討した結果、両境界を統一するプロトコルを選定する。

## Prerequisites

* フロントエンドは React (Vite)、BFF は Fastify (TypeScript) を使用する (ADR-0001)
* バックエンド API は Go (Echo) を使用する (ADR-0001)
* フロント / BFF / バックエンド API の 3 層構成とする (ADR-0001)
* 学習目的のプロジェクトである

## Decision Drivers

* ローカル開発での簡便さ — セットアップの手軽さ、コード生成の要否、デバッグのしやすさ（curl/ブラウザでの動作確認等）
* BFF 集約パターンとの相性 — BFF が複数バックエンド API を集約・変換する際の実装しやすさ
* 型安全性 — フロント↔BFF 間および BFF↔バックエンド間での型の一貫性、TypeScript/Go 間の型同期
* 学習効果と実務応用 — 各プロトコルの概念・設計思想の理解がどの程度実務に活きるか

## Considered Options

* REST + OpenAPI（両境界統一）
* tRPC（FE↔BFF）+ REST + OpenAPI（BFF↔Backend）
* tRPC（FE↔BFF）+ Connect RPC（BFF↔Backend）
* Connect RPC（両境界統一）

## Pros and Cons of the Options

### REST + OpenAPI（両境界統一）

両境界で REST API を使用し、OpenAPI スキーマからのコード生成で TypeScript / Go 間の型同期を実現する構成。

* Good, because curl / Swagger UI / ブラウザで即座にデバッグでき、ローカル開発の簡便さが最も高い
* Good, because REST + OpenAPI は業界で最も広く使われており、学んだスキルがそのまま実務に直結する
* Good, because Fastify の `@fastify/swagger` + TypeBox でコードファーストに OpenAPI スペックを自動生成できる
* Bad, because 2つの境界に対して別々の OpenAPI スペックと codegen パイプラインの管理が必要
* Bad, because OpenAPI の型システムは TypeScript より表現力が低く、codegen を再実行しないと型が古くなるリスクがある

### tRPC（FE↔BFF）+ REST + OpenAPI（BFF↔Backend）

フロント↔BFF 間は tRPC でコード生成不要の型安全を実現し、BFF↔バックエンド間は REST + OpenAPI で接続する分割構成。

* Good, because FE↔BFF 間は tRPC の自動型推論でコード生成不要、最高の開発体験
* Good, because tRPC procedure 内で複数の REST バックエンドを自然に集約・変換でき、BFF パターンに最適
* Good, because 「ゼロ codegen（tRPC）」と「スキーマ駆動（OpenAPI）」の2つのパラダイムを同一プロジェクトで学べる
* Bad, because BFF↔Backend 間は OpenAPI codegen が必要で、FE↔BFF 間とは異なるツールチェーンを管理する必要がある
* Bad, because tRPC の型安全は TypeScript 同士でのみ有効で、BFF 内の tRPC↔REST 型変換層は手動実装

### tRPC（FE↔BFF）+ Connect RPC（BFF↔Backend）

フロント↔BFF 間は tRPC、BFF↔バックエンド間は Connect RPC（Protobuf ベース、gRPC 互換）で型安全を実現する分割構成。

* Good, because FE↔BFF 間は tRPC のゼロ codegen 型推論、BFF↔Backend 間は Protobuf 生成型で両境界とも型安全
* Good, because 「ゼロ codegen（tRPC）」と「スキーマ駆動（Protobuf）」の対照的なアプローチを体験できる
* Bad, because tRPC + Protobuf + buf CLI + 2言語 codegen と、ツールチェーンが最も複雑で初期セットアップが重い
* Bad, because BFF 内で Protobuf メッセージ型→tRPC レスポンス型の変換レイヤーが必要
* Bad, because ミドルウェア / インターセプタの管理が tRPC 側と Connect RPC 側で二重になる

### Connect RPC（両境界統一）

Protobuf スキーマから TypeScript / Go のコードを自動生成し、両境界で統一的な型安全を実現する構成。gRPC 互換だがブラウザから直接呼び出し可能。

* Good, because Protobuf スキーマが唯一の型定義源となり、TypeScript / Go の両方のコードを自動生成でき、真の E2E 型安全を実現
* Good, because Fastify が Connect RPC サーバー（フロントからの受信）兼クライアント（バックエンドへの送信）として自然に機能し、BFF 集約パターンに最適
* Good, because Connect プロトコルの JSON モードにより curl やブラウザ DevTools で直接デバッグ可能
* Good, because Protobuf / gRPC エコシステム（CNCF Sandbox）は業界で広く採用されており、スキーマファースト API 設計の深い理解が得られる
* Bad, because Protobuf + buf CLI の初期セットアップとコード生成ワークフローに学習コストがかかる
* Bad, because ブラウザからのクライアントストリーミング / 双方向ストリーミングは Fetch API の制約で使用不可
* Bad, because Connect RPC の学習リソースは REST / tRPC と比べるとまだ少ない

## Decision Outcome

Chosen option: "Connect RPC（両境界統一）", because 型安全性が高く、今後実務でも利用していきたいため。

### Consequences

* Good, because Protobuf スキーマを唯一の型定義源として TypeScript / Go のコードを自動生成でき、全境界で型安全な通信を実現できる
* Good, because Protobuf / gRPC エコシステム（CNCF）のスキルが身につき、今後の実務にも直接活かせる
* Bad, because Protobuf + buf CLI のセットアップとコード生成ワークフローの習得に初期コストがかかる
* Bad, because Connect RPC の学習リソースは REST / tRPC に比べて少なく、問題解決に時間がかかる場合がある

## More Information

* Connect RPC 公式ドキュメント: https://connectrpc.com/docs
* connect-go (Go ライブラリ): https://github.com/connectrpc/connect-go
* connect-es (TypeScript ライブラリ): https://github.com/connectrpc/connect-es
* @connectrpc/connect-fastify (Fastify プラグイン): https://www.npmjs.com/package/@connectrpc/connect-fastify
* buf CLI (Protobuf 管理): https://buf.build/docs
* フロント↔BFF 間は `@connectrpc/connect-web`、BFF↔バックエンド間は `@connectrpc/connect-node` のトランスポートを使用する
* Go Echo との統合は `echo.WrapHandler()` で Connect RPC ハンドラをマウントする
