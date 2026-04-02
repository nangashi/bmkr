---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# サービス間認証方式の選定

## Decision Outcome

OAuth 2.0 Client Credentials Grant を採用する。

学習価値（実務での汎用性）を最重視した。OAuth 2.0 は業界標準のプロトコルであり、Client Credentials フロー・スコープ設計・トークン検証の実装を通じて得られる知識は、Auth0/Okta/Azure AD 等のマネージドサービスを利用する実務にも直接転用できる。トークン設計の柔軟性も高く、JWT ベースのクレーム設計やステートレス検証の概念も同時に習得できる。

### Accepted Tradeoffs

- 認可サーバ（Keycloak等）の導入・設定コストを受け入れる。Docker Compose への追加と初期設定に時間を要するが、学習投資として許容する
- 認可サーバが単一障害点となるリスクを受け入れる。ローカル環境のみの運用であり、可用性要件は低いため実質的な影響は限定的

### Consequences

- OAuth 2.0 の Client Credentials フロー・スコープ設計・トークン検証を実践的に学べる
- JWT ベースのトークン検証パターンも同時に習得できる
- 認可サーバの障害が全サービス間通信に影響する単一障害点となる
- 認可サーバの設定・管理という本来のスコープ外の学習コストが発生する

## Context and Problem Statement

BFF（Fastify）からバックエンドAPI（Go/Echo）、およびバックエンドAPI間（ECサイト → 商品管理、ECサイト → 顧客管理）の Connect RPC 通信において、リクエスト元が正当なサービスであることを検証する認証方式を決定する必要がある。対象はサービス間リクエストの認証・検証方式であり、エンドユーザー認証（ログイン/パスワード）、管理画面の認証、認可（どのサービスがどのAPIを呼べるかの制御）はスコープ外とする。

## Prerequisites

- BFF: Fastify (TypeScript)、バックエンド: Go (Echo) (ADR-0001)
- サービス間通信: Connect RPC (ADR-0003)
- ローカル環境での実行のみ、学習目的のプロジェクト

## Decision Drivers

- Connect RPC インターセプタとの親和性
- セットアップの手軽さ
- 学習価値（実務での汎用性）
- トークン設計の柔軟性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| 静的 API Key | Pre-Shared Key をヘッダーで送る最小構成 |
| JWT (HMAC-SHA256) | 対称鍵 JWT をインターセプタで付与・検証 |
| mTLS | 相互 TLS 認証によるトランスポート層での認証 |
| **OAuth 2.0 Client Credentials Grant（採用）** | 認可サーバから Bearer トークンを取得しインターセプタで検証 |

除外: SPIFFE/SPIRE（Kubernetes 前提）、Service Mesh（Kubernetes 前提）、OPA（認可エンジンであり認証方式ではない）、クラウド IAM（ローカル不可）、HTTP Basic 認証（gRPC/Connect RPC の慣習から外れる）

## Comparison Overview

| 判断軸 | 静的 API Key | JWT (HMAC-SHA256) | mTLS | OAuth 2.0 Client Credentials |
|--------|-------------|-------------------|------|------------------------------|
| Connect RPC インターセプタとの親和性 | ◎ 公式サンプルそのもの | ◎ ヘッダ読み書きのみで完結 | △ インターセプタ外のTLS設定が必須 | ○ Bearer トークンで完結するが認可サーバが別途必要 |
| セットアップの手軽さ | ◎ 追加依存ゼロ、環境変数1つ | ○ ライブラリ各1つ+共有鍵1つ | △ CA・証明書生成、全サービスへの配布が必要 | △ 認可サーバの構築・設定が必要 |
| 学習価値（実務での汎用性） | △ インターセプタの基本のみ | ◎ JWT・署名検証・クレーム設計等、実務直結 | ○ PKI・TLS・ゼロトラストの基盤知識 | ◎ OAuth 2.0は業界標準、Auth0/Okta等に転用可能 |
| トークン設計の柔軟性 | △ 静的文字列、メタデータ不可 | ◎ iss/sub/aud/exp+カスタムクレーム自由 | △ 証明書のCN/SANのみ、動的クレーム不可 | ◎ JWTベースでスコープ・クレーム自由設計 |

## Notes

- **静的 API Key は有効期限がない**: 鍵漏洩時に手動ローテーションまで無期限に有効となる
- **JWT (HMAC-SHA256) の対称鍵リスク**: 1つ漏洩で全サービスが危殆化する（本番では非対称鍵に移行すべき）
- **mTLS のクレーム制約**: 証明書の CN/SAN のみで動的クレーム（ユーザーID、スコープ等）は持たせられない

## More Information

- [Connect RPC Go Interceptors](https://connectrpc.com/docs/go/interceptors/)
- [Connect RPC Node.js Interceptors](https://connectrpc.com/docs/node/interceptors/)
- [Keycloak Getting Started (Docker)](https://www.keycloak.org/getting-started/getting-started-docker)
- [openid-client v6 (TypeScript)](https://github.com/panva/openid-client)
- [coreos/go-oidc v3 (Go)](https://github.com/coreos/go-oidc)
- BFF→Backend 間の `Authorization` ヘッダーはサービストークン（Client Credentials）専用とする。ユーザー識別のための JWT は `X-User-Token` ヘッダーで別途伝播する（ADR-0013）
