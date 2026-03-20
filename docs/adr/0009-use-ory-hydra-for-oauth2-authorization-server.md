---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# サービス間認証の認可サーバに Ory Hydra を使用する

## Context and Problem Statement

ADR-0006 で OAuth 2.0 Client Credentials Grant をサービス間認証方式として採用した。この決定を実現するための認可サーバ実装を選定する必要がある。Docker Compose でローカル起動可能な OSS であることが必須条件である。この ADR ではエンドユーザー認証（ADR-0005 で決定済み）および認可ポリシーの詳細設計は扱わない。

## Prerequisites

* ADR-0006 で OAuth 2.0 Client Credentials Grant を採用済み
* ADR-0006 で「認可サーバ（Keycloak等）の導入・設定が必要」と明記
* Docker Compose でローカル実行（要件定義）
* BFF は Fastify (TypeScript)、バックエンド API は Go (Echo) を採用済み (ADR-0001)
* サービス間通信は Connect RPC (ADR-0003)
* エンドユーザー認証は ADR-0005 で決定済み（スコープ外）
* 学習目的のプロジェクト

## Decision Drivers

* Client Credentials Grant のセットアップ容易性 — クライアント登録からトークン発行までの手順の少なさ、設定の直感性
* リソース消費量（メモリ・CPU） — ローカル Docker Compose 環境で他サービスと共存するための軽量さ
* Go / TypeScript クライアントライブラリの充実度 — 標準ライブラリや公式 SDK の有無、Connect RPC インターセプタとの統合のしやすさ
* 学習の転用性（実務での遭遇頻度） — 実務のプロジェクトでどれだけ広く使われているか、学んだ知識がどの程度転用可能か

## Considered Options

* Keycloak
* Ory Hydra
* Authentik

### Excluded Options

* Dex — Client Credentials Grant を公式にサポートしておらず、実装予定もないと明言されている
* ZITADEL — 2025年3月にライセンスが Apache-2.0 から AGPL-3.0 に変更された
* Authelia — リバースプロキシ統合が主目的であり、Client Credentials Grant は v4.39.0 で追加された副次的機能
* FusionAuth — コアエンジンがクローズドソース（独自商用ライセンス）であり OSS 要件に不適合

## Comparison Overview

| 判断軸 | Keycloak | Ory Hydra | Authentik |
|--------|----------|-----------|-----------|
| Client Credentials セットアップ容易性 | ○ GUI 3-4ステップ、10-15分。設定項目が多く初見では迷う可能性 | ◎ CLI 3ステップ、10-15分。in-memory モードで DB 不要。Login/Consent App も不要 | ○ Web UI 5-6ステップ。サービスアカウント自動作成が便利だが Flow/Stage 等の独自概念の理解が必要 |
| リソース消費量 | △ アイドル時 600-700MB、最低 750MB。JVM 起因で重い | ◎ Go バイナリ数十MB。in-memory なら Hydra 単体で完結。PostgreSQL 含めても 200-300MB | △ server + worker + PostgreSQL で 1.5-2.5GB。最低要件 2CPU/2GB RAM |
| Go/TS ライブラリ充実度 | ◎ gocloak (1.2k stars)、echo-keycloak、fastify-keycloak-adapter。connectrpc/authn で統合可 | ○ 公式 SDK (Go/TS) あり。標準 golang.org/x/oauth2 でも連携可能。SDK は自動生成でサンプル不足 | ○ 公式 API クライアントあり。標準 OIDC ライブラリで連携可能。Go クライアント Stars 23 と小規模 |
| 学習の転用性 | ◎ OSS IAM のデファクト。日本求人約100件。Red Hat 支援、CNCF Incubating | ○ OpenAI が採用。ThoughtWorks Radar 掲載 (Assess)。Go 製でソースコード読解が容易 | △ GitHub Stars 20k で成長中だがエンタープライズ採用少。日本語情報がほぼない |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Keycloak

* Good, because Admin Console (GUI) でクライアント登録からトークン発行まで視覚的に操作でき、設定を確認しやすい
* Good, because gocloak (1.2k stars)、echo-keycloak など Go エコシステムが充実。connectrpc/authn で Connect RPC とも統合可能
* Good, because OSS IAM のデファクトスタンダード。日本国内でも求人約100件。学んだ概念が Auth0/Okta/Azure AD にそのまま転用可能
* Good, because CNCF Incubating プロジェクトで長期的なメンテナンスが期待でき、Realm Export/Import で設定のバージョン管理も可能
* Bad, because JVM ベースでアイドル時 600-700MB 消費。他サービスと共存する Docker Compose 環境では圧迫が大きい
* Bad, because Java/Quarkus ベースのため内部拡張には Java 知識が必要。Go/TS のみのスキルセットではカスタマイズが困難
* Bad, because GUI の設定項目が多く、Client Credentials に不要な項目も表示されるため初見では迷いやすい

### Ory Hydra

* Good, because Client Credentials Grant は Hydra 単体で完結し、Login/Consent App や追加 UI の実装が不要
* Good, because Go シングルバイナリで数十 MB。in-memory モード (`DSN=memory`) なら DB すら不要で、Docker Compose 環境への負荷が最小
* Good, because OAuth 2.0 仕様に忠実な実装で、仕様理解の学習に最適。JWT モード設定で introspection 不要の高速検証も可能
* Good, because Go 製のためバックエンド (Go/Echo) との技術スタックが一致し、ソースコード読解やデバッグが容易
* Bad, because 市場シェアが Keycloak の半分以下。日本国内での採用事例が限定的で、実務で遭遇する頻度は低い
* Bad, because SDK は OpenAPI 自動生成で実用的なサンプルコードが不足。npm 週間 DL 数も 10,000-23,000 程度
* Bad, because ドキュメントが Ory Network (SaaS) 前提で書かれている箇所があり、セルフホスト情報を探すのに手間がかかることがある

### Authentik

* Good, because サービスアカウントの自動作成機能があり、client_secret を渡すだけで M2M 認証が動作する
* Good, because MIT ライセンスで、Blueprint 機能により設定を YAML でコード管理できる (IaC 的アプローチ)
* Good, because 2025.10 で Redis 依存を除去するなど、アーキテクチャ簡素化に積極的。開発の方向性が良い
* Bad, because server + worker + PostgreSQL で合計 1.5-2.5GB のメモリを消費。3候補中最も重い
* Bad, because Flow/Stage/Policy/Blueprint 等の独自概念の学習曲線が高く、Client Credentials だけ使いたい場合にオーバースペック
* Bad, because エンタープライズ採用が Keycloak 比で少なく、日本語情報もほぼない。学習の転用性が最も低い

## Decision Outcome

Chosen option: "Ory Hydra"

### Rationale

リソース消費量（メモリ・CPU）を最重視した。Go シングルバイナリで数十 MB と圧倒的に軽量であり、in-memory モードなら DB 不要で Docker Compose 環境への負荷が最小である。BFF (Fastify) + バックエンド API (Go/Echo) + Connect RPC + PostgreSQL 等の複数サービスが同一 Docker Compose で共存するローカル環境において、認可サーバがリソースを大きく占有しないことを優先した。Client Credentials Grant が Login/Consent App 不要で Hydra 単体で完結する点も、セットアップ容易性の面で評価した。

### Accepted Tradeoffs

* 学習の転用性が Keycloak と比較して低い（Comparison Overview の学習の転用性「○」に対応）。市場シェアが Keycloak の半分以下であり、日本国内での実務遭遇頻度は限定的。ただし OAuth 2.0 仕様に忠実な実装であるため、プロトコル自体の知識は他の IdP にも転用可能
* Go/TS クライアントライブラリが Keycloak ほど充実していない（Comparison Overview の Go/TS ライブラリ充実度「○」に対応）。SDK は自動生成でサンプルコードが不足しているが、標準 OAuth 2.0 ライブラリ (`golang.org/x/oauth2`, `openid-client`) での連携は可能

### Consequences

* Good, because Go バイナリ数十 MB の軽量な認可サーバで Docker Compose 環境の他サービスを圧迫しない
* Good, because OAuth 2.0 仕様に忠実な実装を通じてプロトコルの理解を深められる
* Bad, because Keycloak と比べて SDK やコミュニティのサンプルコードが少なく、実装時に公式ドキュメント依存になる
* Bad, because Ory Network (SaaS) 前提のドキュメントが混在し、セルフホスト向け情報の取捨選択が必要になる

## More Information

* [Ory Hydra GitHub](https://github.com/ory/hydra)
* [5-minute tutorial](https://www.ory.sh/docs/hydra/5min-tutorial)
* [Client Credentials flow with Ory Hydra](https://www.naiyerasif.com/post/2022/08/21/client-credentials-flow-with-ory-hydra/)
* [Ory Hydra SDK overview](https://www.ory.com/docs/hydra/sdk/overview)
* [connectrpc/authn-go](https://github.com/connectrpc/authn-go)
* [Opaque and JWT access tokens](https://www.ory.com/docs/oauth2-oidc/jwt-access-token)
* [Docker and Deployment Configurations](https://www.ory.com/docs/hydra/self-hosted/configure-deploy)
