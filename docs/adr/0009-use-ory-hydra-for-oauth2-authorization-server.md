---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# サービス間認証の認可サーバの選定

## Decision Outcome

Ory Hydra を採用する。

リソース消費量（メモリ・CPU）を最重視した。Go シングルバイナリで数十 MB と圧倒的に軽量であり、in-memory モードなら DB 不要で Docker Compose 環境への負荷が最小である。BFF (Fastify) + バックエンド API (Go/Echo) + Connect RPC + PostgreSQL 等の複数サービスが同一 Docker Compose で共存するローカル環境において、認可サーバがリソースを大きく占有しないことを優先した。Client Credentials Grant が Login/Consent App 不要で Hydra 単体で完結する点も、セットアップ容易性の面で評価した。

### Accepted Tradeoffs

- 学習の転用性が Keycloak と比較して低い。市場シェアが Keycloak より小さく、日本国内での実務遭遇頻度は限定的。ただし OAuth 2.0 仕様に忠実な実装であるため、プロトコル自体の知識は他の IdP にも転用可能
- Go/TS クライアントライブラリが Keycloak ほど充実していない。SDK は自動生成でサンプルコードが不足しているが、標準 OAuth 2.0 ライブラリ (`golang.org/x/oauth2`, `openid-client`) での連携は可能

### Consequences

- Go バイナリ数十 MB の軽量な認可サーバで Docker Compose 環境の他サービスを圧迫しない
- OAuth 2.0 仕様に忠実な実装を通じてプロトコルの理解を深められる
- Keycloak と比べて SDK やコミュニティのサンプルコードが少なく、実装時に公式ドキュメント依存になる
- Ory Network (SaaS) 前提のドキュメントが混在し、セルフホスト向け情報の取捨選択が必要になる

## Context and Problem Statement

ADR-0006 で OAuth 2.0 Client Credentials Grant をサービス間認証方式として採用した。この決定を実現するための認可サーバ実装を選定する必要がある。Docker Compose でローカル起動可能な OSS であることが必須条件である。エンドユーザー認証（ADR-0005 で決定済み）および認可ポリシーの詳細設計は扱わない。

## Prerequisites

- サービス間認証: OAuth 2.0 Client Credentials Grant (ADR-0006)
- ADR-0006 で「認可サーバ（Keycloak等）の導入・設定が必要」と明記
- Docker Compose でローカル実行、学習目的
- BFF: Fastify (TypeScript)、バックエンド: Go (Echo) (ADR-0001)、サービス間通信: Connect RPC (ADR-0003)

## Decision Drivers

- Client Credentials Grant のセットアップ容易性
- リソース消費量（メモリ・CPU）
- Go / TypeScript クライアントライブラリの充実度
- 学習の転用性（実務での遭遇頻度）

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| Keycloak | JVM ベースの OSS IAM デファクトスタンダード。GUI 管理コンソール付き |
| **Ory Hydra（採用）** | Go 製の軽量 OAuth 2.0 / OIDC サーバ。in-memory モード対応 |
| Authentik | Python 製の統合認証プラットフォーム。Blueprint による IaC 対応 |

除外: Dex（Client Credentials Grant 非サポート）、ZITADEL（AGPL-3.0 に変更）、Authelia（Client Credentials は副次的機能）、FusionAuth（クローズドソース）

## Comparison Overview

| 判断軸 | Keycloak | Ory Hydra | Authentik |
|--------|----------|-----------|-----------|
| Client Credentials セットアップ容易性 | ○ GUI 3-4ステップ。設定項目が多く初見では迷う可能性 | ◎ CLI 3ステップ。in-memory モードで DB 不要。Login/Consent App も不要 | ○ Web UI 5-6ステップ。Flow/Stage 等の独自概念の理解が必要 |
| リソース消費量 | △ アイドル時 600-700MB。JVM 起因で重い | ◎ Go バイナリ数十MB。in-memory なら Hydra 単体で完結 | △ server + worker + PostgreSQL で 1.5-2.5GB |
| Go/TS ライブラリ充実度 | ◎ gocloak、echo-keycloak、fastify-keycloak-adapter。connectrpc/authn で統合可 | ○ 公式 SDK (Go/TS) あり。標準 golang.org/x/oauth2 でも連携可能。SDK は自動生成でサンプル不足 | ○ 公式 API クライアントあり。標準 OIDC ライブラリで連携可能 |
| 学習の転用性 | ◎ OSS IAM のデファクト。Red Hat 支援、CNCF Incubating | ○ OpenAI が採用。ThoughtWorks Radar 掲載 (Assess)。Go 製でソースコード読解が容易 | △ エンタープライズ採用少。日本語情報がほぼない |

## Notes

- **Keycloak の JVM メモリ消費**: アイドル時 600-700MB、最低 750MB で Docker Compose 環境への圧迫が大きい
- **Authentik の高リソース要件**: server + worker + PostgreSQL で合計 1.5-2.5GB。最低要件 2CPU/2GB RAM
- **Authentik の独自概念**: Flow/Stage/Policy/Blueprint 等の学習曲線が高く、Client Credentials だけ使いたい場合にオーバースペック
- **ZITADEL のライセンス変更**: 2025年3月に Apache-2.0 から AGPL-3.0 に変更された

## More Information

- [Ory Hydra GitHub](https://github.com/ory/hydra)
- [5-minute tutorial](https://www.ory.sh/docs/hydra/5min-tutorial)
- [Client Credentials flow with Ory Hydra](https://www.naiyerasif.com/post/2022/08/21/client-credentials-flow-with-ory-hydra/)
- [Ory Hydra SDK overview](https://www.ory.com/docs/hydra/sdk/overview)
- [connectrpc/authn-go](https://github.com/connectrpc/authn-go)
- [Opaque and JWT access tokens](https://www.ory.com/docs/oauth2-oidc/jwt-access-token)
- [Docker and Deployment Configurations](https://www.ory.com/docs/hydra/self-hosted/configure-deploy)
