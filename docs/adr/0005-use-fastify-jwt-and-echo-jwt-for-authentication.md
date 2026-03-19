---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# BFF と管理画面の認証に JWT 自前実装（@fastify/jwt + echo-jwt）を採用する

## Context and Problem Statement

学習用模擬ECサイトにおいて、ECサイト（一般ユーザー）の認証は BFF (Fastify) で管理し、商品管理・顧客管理の管理画面（管理者）の認証は各 Go (Echo) サービスが直接担う。ID/パスワード認証のみを対象とし、BFF での認証状態管理方式（セッション/JWT）およびライブラリを選定する必要がある。認可（RBAC等）の詳細設計およびユーザー登録フローの画面設計は本 ADR のスコープ外とする。

## Prerequisites

* BFF は Fastify (TypeScript) を使用する (ADR-0001)
* バックエンド API は Go (Echo) を使用する (ADR-0001)
* サービス間通信は Connect RPC (Protobuf) を使用する (ADR-0003)
* データベースは PostgreSQL を使用する (ADR-0004)
* ECサイトは ID/パスワード認証、BFF 経由で認証状態を管理する (要件定義)
* 商品管理・顧客管理の管理画面も ID/パスワード認証 (要件定義)
* 学習目的のプロジェクトであり、ローカル実行のみ

## Decision Drivers

* 認証メカニズムの学習深度 — パスワードハッシュ・トークン生成/検証・セッション管理など認証の内部実装をどれだけ低レベルから理解できるか
* Fastify / Echo プラグインとの統合の自然さ — 各フレームワークの設計思想・プラグインシステムに沿って無理なく統合できるか
* Connect RPC での認証情報伝播 — BFF→バックエンド間の Connect RPC 通信で、認証済みユーザー情報をどう安全に伝えるか
* BFF と Go 管理画面への展開の一貫性 — ECサイト認証 (BFF) と管理画面認証 (Go) で類似パターンを適用できるか、学習した知識が相互に活きるか

## Considered Options

* @fastify/jwt + @fastify/cookie (BFF) & echo-jwt + golang-jwt (Go)
* Better Auth (BFF) & echo-jwt + golang-jwt (Go)
* @fastify/secure-session (BFF) & alexedwards/scs (Go)

### Excluded Options

* Lucia Auth — 2025年3月に正式非推奨化。学習リソースに転換済み
* Passport.js — Express 向け設計で Fastify との相性が悪く、OAuth 向けのため ID/パスワード認証のみには過剰
* Auth0 / Clerk / WorkOS — マネージドサービスのためローカル実行不可
* SuperTokens / Ory (Kratos+Hydra) — 別途サービス運用が必要で学習用途にはアーキテクチャが過剰

## Comparison Overview

| 判断軸 | @fastify/jwt + echo-jwt (JWT自前実装) | Better Auth + echo-jwt | @fastify/secure-session + scs (セッション統一) |
|--------|---------|---------|---------|
| 認証メカニズムの学習深度 | ◎ JWT構造・ハッシュ・リフレッシュ戦略を全て自前設計 | △ BFF側はブラックボックス、Go側でのみJWT学習可能 | ○ 暗号化Cookie vs サーバーサイドの2方式を比較学習可能。JWTは学べない |
| Fastify/Echo 統合の自然さ | ◎ 両方とも公式プラグインでネイティブ統合 | △ Better AuthはFetch APIアダプタ経由で非ネイティブ。Echo側はOK | ○ BFF側は公式プラグインで完璧。Go側はscs+Echo互換性に懸念 |
| Connect RPC での認証情報伝播 | ◎ X-User-Token ヘッダーでJWTを転送（ADR-0013）。authn-goで検証 | ○ JWTプラグイン追加で可能だが統合コード自前実装が必要 | △ セッション→ヘッダー変換レイヤーが必要。実装量が最も多い |
| BFF と Go 管理画面への展開の一貫性 | ◎ JWT+bcryptの共通パターンをTS/Goで統一 | △ BFF(高抽象度)とGo(低抽象度)で根本的に異なる | ○ 「セッション」概念は統一だが実装方式が異なる |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### @fastify/jwt + @fastify/cookie (BFF) & echo-jwt + golang-jwt (Go)

Fastify / Echo の公式プラグインで JWT を HttpOnly Cookie に格納する自前実装。両レイヤーで JWT + bcrypt のアプローチを統一し、認証の仕組みを低レベルから学ぶ構成。

* Good, because JWT 構造（Header/Payload/Signature）、パスワードハッシュ、リフレッシュトークン戦略、Cookie セキュリティ属性を全て自分で設計・実装するため、認証メカニズムの学習深度が最も高い
* Good, because @fastify/jwt は Fastify のデコレータ・フック、echo-jwt は Echo のミドルウェアチェーンと、それぞれ公式プラグインとしてネイティブに統合される
* Good, because JWT を X-User-Token ヘッダーで Connect RPC に転送でき（ADR-0013）、authn-go で検証する確立されたパターンがある
* Good, because JWT + bcrypt という共通コンセプトを TypeScript / Go の両方で実装し、知識が相互に活きる
* Good, because 全て permissive ライセンス（MIT/Apache2.0/BSD-3）で外部サービス依存なし
* Bad, because JWT 固有のセキュリティリスクがあり、トークン失効・リフレッシュ・CSRF 対策を全て自前で正しく実装する必要がある
* Bad, because セッション管理・CSRF 対策・トークンローテーションなどのセキュリティ対策コードが全て自己責任となり、実装量が多い

### Better Auth (BFF) & echo-jwt + golang-jwt (Go)

BFF 側は急成長中の TypeScript 認証フレームワーク Better Auth (27K+ Stars) に委譲し、Go 管理画面は JWT 自前実装とする構成。

* Good, because `emailAndPassword: { enabled: true }` の1行で ID/パスワード認証が完成し、実装速度が最も速い
* Good, because Better Auth は 27K+ Stars で活発にメンテナンスされ、セキュリティベストプラクティスがビルトイン
* Good, because Go 管理画面側では JWT 自前実装により、「フレームワーク利用」と「自前実装」の両方を体験できる
* Bad, because BFF 側の認証実装がほぼ全てブラックボックス化され、学習目的のプロジェクトとしては認証の内部理解が浅くなる
* Bad, because Fastify 統合が Fetch API アダプタ経由の非ネイティブ方式で、Fastify の Type Provider / JSON Schema とは独立した世界で動作する
* Bad, because ADR-0004 の「BFF は DB を持たない」前提と矛盾し、アーキテクチャの見直しが必要
* Bad, because SemVer 非準拠の破壊的変更が報告されている（v1.3.25, v1.4.4→v1.4.5）

### @fastify/secure-session (BFF) & alexedwards/scs (Go)

JWT を使わず、BFF は libsodium ベースの暗号化 Cookie セッション、Go は OWASP 準拠のサーバーサイドセッション（PostgreSQL ストア）で統一するセッションベース構成。

* Good, because 暗号化 Cookie セッション（ステートレス）とサーバーサイドセッション（ステートフル）の2方式を同一プロジェクトで比較学習できる
* Good, because CSRF 対策の実装が必須となり、Web セキュリティの実践的な学習機会が得られる
* Good, because scs は OWASP 準拠を明示しており、セキュリティベストプラクティスを学べる
* Bad, because scs は `net/http` ベースで Echo との直接統合に問題があり、echo-scs-session (Stars 21) のメンテナンス継続性に懸念
* Bad, because セッション Cookie は BFF 内で完結するため、Connect RPC での認証情報伝播に追加の変換レイヤーが必要
* Bad, because sodium-native は C ネイティブモジュールのため、Docker (Alpine Linux) 環境でビルド問題が発生する可能性がある

## Decision Outcome

Chosen option: "@fastify/jwt + @fastify/cookie (BFF) & echo-jwt + golang-jwt (Go)", because Fastify / Echo の公式プラグインとしてネイティブに統合でき、JWT + bcrypt の共通パターンを TypeScript / Go 双方で実装することで認証メカニズムの学習効果が最も高いため。

### Rationale

統合の自然さと学習効果を重視した。@fastify/jwt と echo-jwt はそれぞれのフレームワークの公式プラグインであり、デコレータ・ミドルウェアといったフレームワーク固有の設計パターンに沿って無理なく統合できる。また JWT の構造設計・パスワードハッシュ・リフレッシュトークン戦略・Cookie セキュリティ属性を全て自前で実装するため、認証の内部メカニズムを最も深く理解できる。Connect RPC との相性も最も良く、X-User-Token ヘッダーでのユーザー JWT 伝播（ADR-0013）は明確なパターンである。

### Accepted Tradeoffs

* JWT のセキュリティ対策（トークン失効・リフレッシュ・CSRF 対策・トークンローテーション）を全て自前で正しく実装する必要があり、実装量と複雑さが増す（Comparison Overview の学習深度 ◎ の裏返し）
* セキュリティ実装のミスがそのままリスクに直結するが、学習目的かつローカル実行のみのため許容する

### Consequences

* Good, because JWT 構造・パスワードハッシュ・リフレッシュトークン戦略・Cookie セキュリティを実践的に学べる
* Good, because X-User-Token ヘッダーで Connect RPC にユーザー JWT を伝播でき（ADR-0013）、サービス間通信の認証パターンも習得できる
* Bad, because トークンローテーション・CSRF 対策・セッション無効化などのセキュリティ実装を自前で行う必要があり、実装ミスがセキュリティリスクに直結する

## More Information

* @fastify/jwt: https://github.com/fastify/fastify-jwt
* @fastify/cookie: https://github.com/fastify/fastify-cookie
* echo-jwt: https://github.com/labstack/echo-jwt
* golang-jwt/jwt: https://github.com/golang-jwt/jwt
* connectrpc/authn-go: https://github.com/connectrpc/authn-go
* BFF は @fastify/jwt でアクセストークン・リフレッシュトークンを生成し、HttpOnly Cookie に格納する。Go 管理画面は echo-jwt ミドルウェアで JWT を検証する。BFF→バックエンド間は Connect RPC の X-User-Token ヘッダーでユーザー JWT を伝播し、Backend で検証する（ADR-0013: Authorization ヘッダーはサービス認証専用）
