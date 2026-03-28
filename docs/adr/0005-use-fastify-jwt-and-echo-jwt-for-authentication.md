---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# BFF・管理画面の認証方式とライブラリの選定

## Decision Outcome

@fastify/jwt + @fastify/cookie (BFF) & echo-jwt + golang-jwt (Go) を採用する。

Fastify / Echo の公式プラグインとしてネイティブに統合でき、JWT + bcrypt の共通パターンを TypeScript / Go 双方で実装することで認証メカニズムの学習効果が最も高い。Connect RPC との相性も良く、X-User-Token ヘッダーでのユーザー JWT 伝播（ADR-0013）は明確なパターンである。

### Accepted Tradeoffs

- JWT のセキュリティ対策（トークン失効・リフレッシュ・CSRF 対策・トークンローテーション）を全て自前で正しく実装する必要があり、実装量と複雑さが増す。ローカル実行のみのため許容する

### Consequences

- JWT 構造・パスワードハッシュ・リフレッシュトークン戦略・Cookie セキュリティを実践的に学べる
- X-User-Token ヘッダーで Connect RPC にユーザー JWT を伝播でき（ADR-0013）、サービス間通信の認証パターンも習得できる
- トークンローテーション・CSRF 対策・セッション無効化などのセキュリティ実装を自前で行う必要があり、実装ミスがセキュリティリスクに直結する

## Context and Problem Statement

ECサイト（一般ユーザー）の認証は BFF (Fastify) で、管理画面（管理者）の認証は各 Go (Echo) サービスが担う。ID/パスワード認証の認証状態管理方式とライブラリを選定する。認可（RBAC等）およびユーザー登録フローの画面設計はスコープ外。

## Prerequisites

- BFF: Fastify (TypeScript)、バックエンド: Go (Echo) (ADR-0001)
- サービス間通信: Connect RPC (ADR-0003)、DB: PostgreSQL (ADR-0004)
- ID/パスワード認証、ローカル実行のみ

## Decision Drivers

- 認証メカニズムの学習深度
- Fastify / Echo プラグインとの統合の自然さ
- Connect RPC での認証情報伝播
- BFF と Go 管理画面への展開の一貫性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **@fastify/jwt + echo-jwt（採用）** | 公式プラグインで JWT を HttpOnly Cookie に格納する自前実装 |
| Better Auth + echo-jwt | BFF 側は認証フレームワークに委譲、Go 側は JWT 自前 |
| @fastify/secure-session + scs | JWT を使わず暗号化 Cookie + サーバーサイドセッション |

除外: Lucia Auth（非推奨化済み）、Passport.js（Express 向け）、Auth0/Clerk/WorkOS（マネージド、ローカル不可）、SuperTokens/Ory Kratos+Hydra（アーキテクチャが過剰）

## Comparison Overview

| 判断軸 | @fastify/jwt + echo-jwt | Better Auth + echo-jwt | secure-session + scs |
|--------|:-:|:-:|:-:|
| 認証メカニズムの学習深度 | ◎ JWT構造・ハッシュ・リフレッシュを全て自前設計 | △ BFF側はブラックボックス | ○ 暗号化Cookie vs サーバーサイドを比較学習可能 |
| Fastify/Echo 統合 | ◎ 両方とも公式プラグイン | △ Fetch APIアダプタ経由で非ネイティブ | ○ BFF側は公式、scs+Echo互換性に懸念 |
| Connect RPC 伝播 | ◎ X-User-Tokenヘッダーで転送 (ADR-0013) | ○ JWTプラグイン追加で可能 | △ セッション→ヘッダー変換レイヤーが必要 |
| BFF/Go 一貫性 | ◎ JWT+bcryptをTS/Goで統一 | △ 抽象度が根本的に異なる | ○ 概念は統一だが実装方式が異なる |

## Notes

- **Better Auth は ADR-0004 の前提と矛盾**: Better Auth は自前 DB（セッション・アカウント管理）を必要とし、「BFF は RDB を持たない」前提と整合しない
- **Better Auth の SemVer 非準拠**: v1.3.25, v1.4.4→v1.4.5 で破壊的変更が報告されている
- **secure-session の sodium-native**: C ネイティブモジュールのため Docker (Alpine Linux) 環境でビルド問題が発生する可能性がある

## More Information

- [@fastify/jwt](https://github.com/fastify/fastify-jwt), [@fastify/cookie](https://github.com/fastify/fastify-cookie)
- [echo-jwt](https://github.com/labstack/echo-jwt), [golang-jwt](https://github.com/golang-jwt/jwt)
- [connectrpc/authn-go](https://github.com/connectrpc/authn-go)
- Authorization ヘッダーはサービス認証専用、ユーザー JWT は X-User-Token で伝播（ADR-0013）
