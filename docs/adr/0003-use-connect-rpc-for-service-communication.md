---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# マイクロサービス間の通信プロトコル選定

## Decision Outcome

フロント↔BFF 間および BFF↔バックエンド間の両境界で Connect RPC を採用する。Protobuf スキーマを唯一の型定義源として TypeScript / Go の両方のコードを自動生成でき、真の E2E 型安全を実現できる。Fastify が Connect RPC サーバー（フロントからの受信）兼クライアント（バックエンドへの送信）として自然に機能し、BFF 集約パターンに最適である。また Connect プロトコルの JSON モードにより curl やブラウザ DevTools で直接デバッグ可能で、ローカル開発の簡便さも確保される。Protobuf / gRPC エコシステム（CNCF Sandbox）は業界で広く採用されており、スキーマファースト API 設計の深い理解が得られ、今後の実務にも直接活かせる。

### Consequences

- Protobuf スキーマを唯一の型定義源として TypeScript / Go のコードを自動生成でき、全境界で型安全な通信を実現できる
- Protobuf / gRPC エコシステム（CNCF）のスキルが身につき、今後の実務にも直接活かせる
- Protobuf + buf CLI のセットアップとコード生成ワークフローの習得に初期コストがかかる
- Connect RPC の学習リソースは REST / tRPC に比べて少なく、問題解決に時間がかかる場合がある

## Context and Problem Statement

React (Vite) + Fastify BFF + Go Echo バックエンド API の 3 層構成において、フロント↔BFF 間および BFF↔バックエンド間の通信プロトコルを選定する必要がある。2つの境界で異なるプロトコルを採用する可能性も含めて検討した結果、両境界を統一するプロトコルを選定する。

## Prerequisites

- フロントエンド: React (Vite)、BFF: Fastify (TypeScript)、バックエンドAPI: Go (Echo) (ADR-0001)
- フロント / BFF / バックエンド API の 3 層構成 (ADR-0001)
- 学習目的のプロジェクトである

## Decision Drivers

- ローカル開発での簡便さ
- BFF 集約パターンとの相性
- 型安全性
- 学習効果と実務応用

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **Connect RPC・両境界統一（採用）** | Protobuf スキーマから TypeScript / Go のコードを自動生成し、両境界で統一的な型安全を実現する構成 |
| REST + OpenAPI（両境界統一） | 両境界で REST API を使用し、OpenAPI スキーマからのコード生成で型同期を実現する構成 |
| tRPC（FE↔BFF）+ REST + OpenAPI（BFF↔Backend） | フロント↔BFF 間は tRPC でコード生成不要の型安全、BFF↔バックエンド間は REST + OpenAPI で接続する分割構成 |
| tRPC（FE↔BFF）+ Connect RPC（BFF↔Backend） | フロント↔BFF 間は tRPC、BFF↔バックエンド間は Connect RPC で型安全を実現する分割構成 |

## Comparison Overview

| 観点 | REST + OpenAPI | tRPC + REST | tRPC + Connect RPC | Connect RPC（統一） |
|------|---------------|-------------|--------------------|--------------------|
| ローカル開発での簡便さ | ◎ | ○ | △ | ○ |
| BFF 集約パターンとの相性 | ○ | ◎ | ○ | ◎ |
| 型安全性 | ○ | ○ | ○ | ◎ |
| 学習効果と実務応用 | ○ | ○ | ○ | ◎ |
| ツールチェーンの単純さ | ○ | △ | △ | ○ |

## Notes

- tRPC の型安全は TypeScript 同士でのみ有効で、Go バックエンドとの型同期には別のメカニズムが必要
- tRPC + Connect RPC の組み合わせはツールチェーンが最も複雑（tRPC + Protobuf + buf CLI + 2言語 codegen）で、BFF 内で Protobuf メッセージ型→tRPC レスポンス型の変換レイヤーも必要
- ブラウザからのクライアントストリーミング / 双方向ストリーミングは Fetch API の制約で使用不可

## More Information

- Connect RPC 公式ドキュメント: https://connectrpc.com/docs
- connect-go (Go ライブラリ): https://github.com/connectrpc/connect-go
- connect-es (TypeScript ライブラリ): https://github.com/connectrpc/connect-es
- @connectrpc/connect-fastify (Fastify プラグイン): https://www.npmjs.com/package/@connectrpc/connect-fastify
- buf CLI (Protobuf 管理): https://buf.build/docs
- フロント↔BFF 間は `@connectrpc/connect-web`、BFF↔バックエンド間は `@connectrpc/connect-node` のトランスポートを使用する
- Go Echo との統合は `echo.WrapHandler()` で Connect RPC ハンドラをマウントする
