---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# サービス間認証に OAuth 2.0 Client Credentials Grant を使用する

## Context and Problem Statement

BFF（Fastify）からバックエンドAPI（Go/Echo）、およびバックエンドAPI間（ECサイト → 商品管理、ECサイト → 顧客管理）の Connect RPC 通信において、リクエスト元が正当なサービスであることを検証する認証方式を決定する必要がある。対象はサービス間リクエストの認証・検証方式であり、エンドユーザー認証（ログイン/パスワード）、管理画面の認証、認可（どのサービスがどのAPIを呼べるかの制御）はスコープ外とする。

## Prerequisites

* BFF は Fastify (TypeScript)、バックエンドAPI は Go (Echo) を採用済み (ADR-0001)
* サービス間通信は Connect RPC を採用済み (ADR-0003)
* ローカル環境での実行のみを想定（要件定義）
* 学習目的のプロジェクト

## Decision Drivers

* Connect RPC インターセプタとの親和性 — インターセプタ機構のみで認証ロジックが完結するか
* セットアップの手軽さ — ローカル環境で動かすまでに必要な設定・依存の少なさ
* 学習価値（実務での汎用性） — この方式を学ぶことで実務のマイクロサービス開発に活かせる知識がどれだけ得られるか
* トークン設計の柔軟性 — 有効期限、発行元サービスID、対象サービスの指定など、認証情報にメタデータを持たせられるか

## Considered Options

* 静的 API Key（Pre-Shared Key）
* JWT（HMAC-SHA256 対称鍵）
* mTLS（相互TLS認証）
* OAuth 2.0 Client Credentials Grant

### Excluded Options

* SPIFFE/SPIRE — Kubernetes/クラウド環境が前提であり、ローカル開発環境では過剰
* Service Mesh (Istio, Linkerd) — Kubernetes が前提であり、構築コストが過大
* OPA (Open Policy Agent) — 認可エンジンであり、認証方式そのものではない
* クラウド IAM (AWS IAM, GCP IAM 等) — ローカル環境では使用不可
* HTTP Basic 認証 — gRPC/Connect RPC の慣習から外れる

## Comparison Overview

| 判断軸 | 静的 API Key | JWT (HMAC-SHA256) | mTLS | OAuth 2.0 Client Credentials |
|--------|-------------|-------------------|------|------------------------------|
| Connect RPC インターセプタとの親和性 | ◎ 公式サンプルそのもの | ◎ ヘッダ読み書きのみで完結 | △ インターセプタ外のTLS設定が必須 | ○ Bearer トークンで完結するが認可サーバが別途必要 |
| セットアップの手軽さ | ◎ 追加依存ゼロ、環境変数1つ | ○ ライブラリ各1つ+共有鍵1つ | △ CA・証明書生成、全サービスへの配布が必要 | △ 認可サーバ(Keycloak等)の構築・設定が必要 |
| 学習価値（実務での汎用性） | △ インターセプタの基本のみ | ◎ JWT・署名検証・クレーム設計等、実務直結 | ○ PKI・TLS・ゼロトラストの基盤知識 | ◎ OAuth 2.0は業界標準、Auth0/Okta等に転用可能 |
| トークン設計の柔軟性 | △ 静的文字列、メタデータ不可 | ◎ iss/sub/aud/exp+カスタムクレーム自由 | △ 証明書のCN/SANのみ、動的クレーム不可 | ◎ JWTベースでスコープ・クレーム自由設計 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### 静的 API Key（Pre-Shared Key）

* Good, because Connect RPC 公式ドキュメントのインターセプタ例がこの方式そのもので、実装が最も容易（Go/TS 各10-20行）
* Good, because 追加依存パッケージゼロ、環境変数に鍵を1つ設定するだけで動作
* Good, because curl で `--header` を付けるだけでテスト可能
* Bad, because トークン設計（有効期限、署名検証、クレーム）等の実務で重要な概念を学べない
* Bad, because 有効期限がなく、鍵漏洩時に手動ローテーションまで無期限に有効
* Bad, because サービス数増加時に鍵管理コストが急増

### JWT（HMAC-SHA256 対称鍵）

* Good, because インターセプタ内でヘッダ読み書きのみで認証ロジックが完結する
* Good, because Go (`golang-jwt/jwt/v5`) と TS (`jose`) 各1パッケージ + 共有鍵1つで動作
* Good, because JWT構造・標準クレーム・署名検証・アルゴリズム混同攻撃対策など実務直結の知識が得られる
* Good, because iss/sub/aud/exp/カスタムクレームを自由に設計でき、ステートレス検証（DB不要）
* Bad, because 対称鍵の場合、1つ漏洩で全サービスが危殆化（本番では非対称鍵に移行すべき）
* Bad, because 発行済みトークンの即時無効化が不可能（短い exp で緩和）

### mTLS（相互TLS認証）

* Good, because PKI・TLSハンドシェイク・ゼロトラストの基盤知識が得られ、サービスメッシュ(Istio等)の原理理解に活きる
* Good, because セキュリティ強度が最も高く、トークン窃取リスクがない
* Bad, because Connect RPC インターセプタのみでは完結せず、TLS設定→ミドルウェア→コンテキスト注入の多層構成が必要
* Bad, because ローカルCA作成、サービスごとの証明書生成・配布が必要でセットアップが重い
* Bad, because 証明書のCN/SANのみで動的クレーム（ユーザーID、スコープ等）は持たせられない

### OAuth 2.0 Client Credentials Grant

* Good, because OAuth 2.0 は業界標準であり、Auth0/Okta/Azure AD 等のマネージドサービスにも知識が転用可能
* Good, because Bearer トークン（JWT）をインターセプタで付与・検証するパターンで完結
* Good, because スコープによる権限制御や JWT クレームの柔軟な設計が可能
* Good, because トークンの自動更新がライブラリに内蔵
* Bad, because 認可サーバ（Keycloak等）の導入・設定が必要で初期構築に30分〜1時間
* Bad, because 2〜3サービスの学習環境ではオーバーエンジニアリングになる可能性
* Bad, because 認可サーバという依存コンポーネントの追加で障害ポイントが増える

## Decision Outcome

Chosen option: "OAuth 2.0 Client Credentials Grant"

### Rationale

学習価値（実務での汎用性）を最重視した。OAuth 2.0 は業界標準のプロトコルであり、Client Credentials フロー・スコープ設計・トークン検証の実装を通じて得られる知識は、Auth0/Okta/Azure AD 等のマネージドサービスを利用する実務にも直接転用できる。トークン設計の柔軟性も高く、JWT ベースのクレーム設計やステートレス検証の概念も同時に習得できる点を評価した。

### Accepted Tradeoffs

* 認可サーバ（Keycloak等）の導入・設定コストを受け入れる（Comparison Overview のセットアップの手軽さ「△」に対応）。Docker Compose への追加と初期設定に30分〜1時間を要するが、学習投資として許容する
* 認可サーバが単一障害点となるリスクを受け入れる。ローカル環境のみの運用であり、可用性要件は低いため実質的な影響は限定的

### Consequences

* Good, because OAuth 2.0 の Client Credentials フロー・スコープ設計・トークン検証を実践的に学べる
* Good, because JWT ベースのトークン検証パターンも同時に習得できる
* Bad, because 認可サーバの障害が全サービス間通信に影響する単一障害点となる
* Bad, because 認可サーバの設定・管理という本来のスコープ外の学習コストが発生する

## More Information

* [Connect RPC Go Interceptors](https://connectrpc.com/docs/go/interceptors/)
* [Connect RPC Node.js Interceptors](https://connectrpc.com/docs/node/interceptors/)
* [Keycloak Getting Started (Docker)](https://www.keycloak.org/getting-started/getting-started-docker)
* [openid-client v6 (TypeScript)](https://github.com/panva/openid-client)
* [coreos/go-oidc v3 (Go)](https://github.com/coreos/go-oidc)
