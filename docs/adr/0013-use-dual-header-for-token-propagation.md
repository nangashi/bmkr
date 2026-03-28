---
status: "accepted"
date: 2026-03-19
last-validated: 2026-03-19
---

# BFF→Backend 間のトークン伝播方式の選定

## Decision Outcome

実装の単純さと既存 ADR との整合性を重視し、二重ヘッダー方式（Authorization + X-User-Token）を採用する。`Authorization` ヘッダーをサービス認証（Client Credentials トークン）専用とし、ユーザー JWT は `X-User-Token` カスタムヘッダーで伝播する。ヘッダー名の修正のみで ADR-0005 と ADR-0006 の衝突を解消でき、既存の認証ライブラリやインターセプタの変更を最小限に抑えられる。Backend の検証順序もサービス認証→ユーザー識別と段階的に明確化され、障害時の切り分けが容易になる。

### ヘッダー設計

| ヘッダー | 用途 | トークン種別 | 検証主体 |
|----------|------|-------------|----------|
| `Authorization: Bearer <token>` | サービス認証 | Client Credentials トークン (ADR-0006) | authn-go / Connect RPC インターセプタ |
| `X-User-Token: <jwt>` | ユーザー識別 | ユーザー JWT (ADR-0005) | Backend のミドルウェア / ハンドラ |

### Backend 検証順序

1. **サービス認証**: `Authorization` ヘッダーの Client Credentials トークンを検証し、リクエスト元が正当なサービスであることを確認する
2. **ユーザー識別**: `X-User-Token` ヘッダーのユーザー JWT を検証し、操作を行うユーザーを識別する

サービス認証が失敗した場合はユーザー識別を行わず、即座にリクエストを拒否する。

### リクエストフロー

```
ブラウザ → BFF (Fastify)
  Cookie: access_token=<user-jwt>
  ↓
  BFF が Cookie からユーザー JWT を取得
  BFF が認可サーバから Client Credentials トークンを取得（キャッシュ済み）
  ↓
BFF → Backend (Go/Echo) [Connect RPC]
  Authorization: Bearer <client-credentials-token>
  X-User-Token: <user-jwt>
  ↓
  Backend: 1) Authorization でサービス認証
  Backend: 2) X-User-Token でユーザー識別
```

### Backend→Backend 間の伝播ルール

* **ユーザー操作起因のリクエスト**: X-User-Token を後続サービスに転送する（例: ECサイト→商品管理で在庫確認）
* **バッチ処理・システム起因のリクエスト**: X-User-Token は不要、Authorization のみで通信する

### Accepted Tradeoffs

- X-User-Token はカスタムヘッダーであり、標準仕様ではない。ただし、サービス内部通信のヘッダー設計はプロジェクト固有であり、外部互換性は不要
- Backend→Backend 間の X-User-Token 転送は各サービスの実装ルールとして遵守する必要がある。Connect RPC インターセプタで共通化することで対応する

### Consequences

- ADR-0005 と ADR-0006 の Authorization ヘッダーの衝突が解消され、各 ADR の責務が明確になる
- Backend の検証順序がサービス認証→ユーザー識別と段階的に定義され、実装・デバッグが容易になる
- X-User-Token のヘッダー転送を各サービスのインターセプタで実装・維持する必要がある

## Context and Problem Statement

ADR-0005（ユーザー認証）では「JWT を Authorization ヘッダーでそのまま Connect RPC に転送」、ADR-0006（サービス間認証）では「Bearer トークンで完結」と規定しており、BFF→Backend 間の `Authorization` ヘッダーの用途が衝突している。サービス認証トークン（Client Credentials）とユーザー識別トークン（ユーザー JWT）を同一ヘッダーで伝播することはできないため、両トークンの伝播方式を明確に分離する必要がある。

## Prerequisites

- BFF→バックエンド間は Connect RPC で通信 (ADR-0003)、ユーザー認証は @fastify/jwt + echo-jwt (ADR-0005)
- サービス間認証は OAuth 2.0 Client Credentials Grant (ADR-0006)、認可サーバは Ory Hydra (ADR-0009)

## Decision Drivers

- ヘッダー用途の一意性
- 実装の単純さ
- 検証順序の明確さ
- 既存 ADR との整合性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| Token Exchange (RFC 8693) | ユーザー JWT とサービストークンを認可サーバで単一トークンに交換する方式 |
| **二重ヘッダー方式（Authorization + X-User-Token）（採用）** | Authorization をサービス認証専用、X-User-Token でユーザー JWT を伝播する方式 |
| BFF 署名付きユーザーコンテキスト | BFF がユーザー情報を独自フォーマットで署名し、カスタムヘッダーで伝播する方式 |

## Comparison Overview

| 判断軸 | Token Exchange (RFC 8693) | 二重ヘッダー方式 | BFF 署名付きユーザーコンテキスト |
|--------|--------------------------|------------------|--------------------------------|
| ヘッダー用途の一意性 | ◎ 単一トークンに統合 | ◎ ヘッダーごとに用途を分離 | ○ Authorization + カスタムヘッダー |
| 実装の単純さ | △ 認可サーバの Token Exchange エンドポイント設定が必要 | ◎ ヘッダー追加のみ、ライブラリ変更不要 | △ BFF での署名生成・Backend での署名検証が追加 |
| 検証順序の明確さ | ○ 単一トークンのクレームで判断 | ◎ ヘッダー単位で段階的に検証可能 | ○ 署名検証が追加ステップとなる |
| 既存 ADR との整合性 | △ ADR-0005/0006 の両方を大幅改訂 | ◎ ヘッダー名の変更のみで既存決定を維持 | △ BFF の責務が拡大し ADR-0005 の改訂が必要 |

## Notes

- Token Exchange (RFC 8693) はリクエストごとにトークン交換の往復通信が発生し、レイテンシが増加する
- BFF 署名付きユーザーコンテキスト方式は BFF に署名鍵の管理と署名生成の責務が追加され、ADR-0005 の「JWT をそのまま伝播」という設計方針から逸脱する

## More Information

* ADR-0005: ユーザー認証（@fastify/jwt + echo-jwt）— ユーザー JWT の生成・検証を規定
* ADR-0006: サービス間認証（OAuth 2.0 Client Credentials）— サービストークンの取得・検証を規定
* ADR-0009: 認可サーバ（Ory Hydra）— Client Credentials トークンの発行元
