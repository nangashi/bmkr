---
status: "accepted"
date: 2026-03-19
last-validated: 2026-03-19
---

# BFF→Backend 間のトークン伝播に二重ヘッダー方式を採用する

## Context and Problem Statement

ADR-0005（ユーザー認証）では「JWT を Authorization ヘッダーでそのまま Connect RPC に転送」、ADR-0006（サービス間認証）では「Bearer トークンで完結」と規定しており、BFF→Backend 間の `Authorization` ヘッダーの用途が衝突している。サービス認証トークン（Client Credentials）とユーザー識別トークン（ユーザー JWT）を同一ヘッダーで伝播することはできないため、両トークンの伝播方式を明確に分離する必要がある。

## Prerequisites

* BFF→バックエンド間は Connect RPC で通信する (ADR-0003)
* ユーザー認証は @fastify/jwt + echo-jwt による JWT 自前実装 (ADR-0005)
* サービス間認証は OAuth 2.0 Client Credentials Grant (ADR-0006)
* 認可サーバは Ory Hydra (ADR-0009)

## Decision Drivers

* ヘッダー用途の一意性 — 1つのヘッダーが1つの目的にのみ使用され、ADR 間で矛盾がないこと
* 実装の単純さ — BFF・Backend 双方の変更量が少なく、既存の認証ライブラリとの統合が容易であること
* 検証順序の明確さ — Backend がリクエストを受け取った際に、サービス認証→ユーザー識別の順序で段階的に検証できること
* 既存 ADR との整合性 — ADR-0005 と ADR-0006 の決定内容を大きく変更せず、補完する形であること

## Considered Options

* Token Exchange (RFC 8693)
* 二重ヘッダー方式（Authorization + X-User-Token）
* BFF 署名付きユーザーコンテキスト

## Comparison Overview

| 判断軸 | Token Exchange (RFC 8693) | 二重ヘッダー方式 | BFF 署名付きユーザーコンテキスト |
|--------|--------------------------|------------------|--------------------------------|
| ヘッダー用途の一意性 | ◎ 単一トークンに統合 | ◎ ヘッダーごとに用途を分離 | ○ Authorization + カスタムヘッダー |
| 実装の単純さ | △ 認可サーバの Token Exchange エンドポイント設定が必要 | ◎ ヘッダー追加のみ、ライブラリ変更不要 | △ BFF での署名生成・Backend での署名検証が追加 |
| 検証順序の明確さ | ○ 単一トークンのクレームで判断 | ◎ ヘッダー単位で段階的に検証可能 | ○ 署名検証が追加ステップとなる |
| 既存 ADR との整合性 | △ ADR-0005/0006 の両方を大幅改訂 | ◎ ヘッダー名の変更のみで既存決定を維持 | △ BFF の責務が拡大し ADR-0005 の改訂が必要 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Token Exchange (RFC 8693)

BFF が認可サーバに対してユーザー JWT とサービストークンを提示し、両方の情報を含む単一のトークンに交換する方式。

* Good, because 単一の Authorization ヘッダーで全ての認証情報を伝播でき、ヘッダー設計がシンプル
* Good, because RFC 標準に準拠しており、実務での汎用性が高い
* Bad, because 認可サーバに Token Exchange エンドポイントの追加設定が必要で、Ory Hydra での対応状況に依存する
* Bad, because リクエストごとにトークン交換の往復通信が発生し、レイテンシが増加する
* Bad, because ADR-0005 と ADR-0006 の両方を大幅に改訂する必要がある

### 二重ヘッダー方式（Authorization + X-User-Token）

`Authorization` ヘッダーをサービス認証（Client Credentials トークン）専用とし、ユーザー JWT は `X-User-Token` カスタムヘッダーで伝播する方式。

* Good, because ヘッダーごとに用途が明確に分離され、ADR-0005 と ADR-0006 の衝突を解消できる
* Good, because Backend はまず Authorization でサービス認証、次に X-User-Token でユーザー識別と、段階的に検証できる
* Good, because 既存の認証ライブラリ（authn-go、echo-jwt）の変更が不要で、カスタムヘッダーの読み取りを追加するだけ
* Good, because ADR-0005 と ADR-0006 の決定内容はそのまま維持し、ヘッダー名の修正のみで整合性を確保できる
* Bad, because X-User-Token はカスタムヘッダーであり、標準化されていない
* Bad, because Backend→Backend 間でユーザーコンテキストを伝播する場合、各サービスが X-User-Token を転送するルールを遵守する必要がある

### BFF 署名付きユーザーコンテキスト

BFF がユーザー情報を独自フォーマットで署名し、カスタムヘッダーで伝播する方式。Authorization はサービストークン、X-User-Context に BFF が署名したユーザー情報を格納する。

* Good, because BFF がユーザー情報のフォーマットを制御でき、必要最小限の情報のみを伝播できる
* Good, because ユーザー JWT の有効期限とは独立した署名を付与できる
* Bad, because BFF に署名鍵の管理と署名生成の責務が追加され、BFF の複雑性が増す
* Bad, because Backend が BFF の署名を検証するための鍵共有メカニズムが追加で必要
* Bad, because ADR-0005 の「JWT をそのまま伝播」という設計方針から逸脱する

## Decision Outcome

Chosen option: "二重ヘッダー方式（Authorization + X-User-Token）"

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

### Rationale

実装の単純さと既存 ADR との整合性を重視した。二重ヘッダー方式はヘッダー名の修正のみで ADR-0005 と ADR-0006 の衝突を解消でき、既存の認証ライブラリやインターセプタの変更を最小限に抑えられる。Backend の検証順序もサービス認証→ユーザー識別と段階的に明確化され、障害時の切り分けが容易になる。

### Accepted Tradeoffs

* X-User-Token はカスタムヘッダーであり、標準仕様ではない。ただし、サービス内部通信のヘッダー設計はプロジェクト固有であり、外部互換性は不要
* Backend→Backend 間の X-User-Token 転送は各サービスの実装ルールとして遵守する必要がある。Connect RPC インターセプタで共通化することで対応する

### Consequences

* Good, because ADR-0005 と ADR-0006 の Authorization ヘッダーの衝突が解消され、各 ADR の責務が明確になる
* Good, because Backend の検証順序がサービス認証→ユーザー識別と段階的に定義され、実装・デバッグが容易になる
* Bad, because X-User-Token のヘッダー転送を各サービスのインターセプタで実装・維持する必要がある

## More Information

* ADR-0005: ユーザー認証（@fastify/jwt + echo-jwt）— ユーザー JWT の生成・検証を規定
* ADR-0006: サービス間認証（OAuth 2.0 Client Credentials）— サービストークンの取得・検証を規定
* ADR-0009: 認可サーバ（Ory Hydra）— Client Credentials トークンの発行元
