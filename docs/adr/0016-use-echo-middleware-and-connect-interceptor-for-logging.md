---
status: "accepted"
date: 2026-03-23
last-validated: 2026-03-23
---

# ログ出力パターンの選定

## Decision Outcome

Echo ミドルウェア + Connect インターセプタの組み合わせを採用する。Connect インターセプタで `Spec().Procedure` と `CodeOf(err)` を使い RPC メタデータを正確に取得しつつ、Echo ミドルウェアで非 RPC エンドポイント（管理画面・ヘルスチェック）もカバーする。Echo ミドルウェア単独ではストリーミング RPC / gRPC プロトコル対応時に破綻するリスクがあり、Connect インターセプタ単独では管理画面のログが欠けて3サービス統一が崩れる。組み合わせにより両方の弱点を補完できる。

### Accepted Tradeoffs

- 2レイヤー管理（Connect インターセプタ + Echo ミドルウェア）と Skipper 保守の複雑さを受け入れる。ログフォーマット変更時に両方の更新が必要になる
- RPC ログと非 RPC ログでフォーマットが完全同一にはならない（RPC: `method: ec.v1.CartService/GetCart`、非 RPC: `method: GET /admin/products`）

### Consequences

- RPC エンドポイントは Connect インターセプタで `Spec().Procedure` と `CodeOf(err)` を使い、canonical log line の4必須フィールドを正確に出力できる
- 非 RPC エンドポイント（管理画面・ヘルスチェック）は Echo ミドルウェアでカバーされ、ログに穴がない
- OTel 導入時に Connect 側は otelconnect、Echo 側は otelecho と各レイヤーの標準ツールを段階的に追加できる

## Context and Problem Statement

log/slog を採用し（ADR-0015）、ログ設計ガイドで canonical log line パターンが定義済みだが、ログをどのレイヤーで出力するかが未決定である。Echo + Connect RPC 構成ではリクエストの流れが `Echo ミドルウェア → echo.WrapHandler → Connect インターセプタ → RPC ハンドラ` となり、各レイヤーでアクセスできる情報が異なる。この ADR ではログ出力の実装パターンを選定する。

## Prerequisites

- log/slog を採用済み (ADR-0015)、バックエンド API は Go (Echo) + Connect RPC (ADR-0001, ADR-0003)
- ログ設計ガイド (docs/guides/go-logging.md) で canonical log line パターンが定義済み — RPC メソッド名、ステータス、duration_ms、request_id を含む1行ログ
- product-mgmt は管理画面（templ + HTMX）も Echo ルートで提供 (ADR-0007)
- 3サービスで同一の実装パターンを使い、将来の OpenTelemetry 統合を妨げないこと

## Decision Drivers

- canonical log line との適合性
- ログカバレッジ
- 実装・保守の複雑さ
- OTel 統合パスとの整合性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| Echo ミドルウェア | `RequestLoggerWithConfig` で全エンドポイントを一括カバー |
| Connect インターセプタ | RPC メタデータをネイティブに取得。非 RPC は未カバー |
| **Echo ミドルウェア + Connect インターセプタの組み合わせ（採用）** | RPC は Connect で正確に取得、非 RPC は Echo でカバー |

除外: ハンドラ直接呼び出し（ログ設計ガイドでアンチパターンとして定義）

## Comparison Overview

| 判断軸 | Echo ミドルウェア | Connect インターセプタ | 組み合わせ |
|--------|-----------------|---------------------|-----------|
| canonical log line との適合性 | △ RPC メソッド名は URI パス推測、ストリーミング/gRPC で破綻 | ◎ 4フィールド全てネイティブ取得 | ◎ RPC は Connect で正確に、非 RPC は Echo でカバー |
| ログカバレッジ | ◎ RPC・非 RPC ともワンストップでカバー | △ 非 RPC（管理画面）が未カバー | ◎ 全エンドポイントをカバー、Skipper で二重出力防止 |
| 実装・保守の複雑さ | ◎ `RequestLoggerWithConfig` の設定だけ | ○ 30-40行の自作、管理画面は別途対応 | ○ 2レイヤー管理 + Skipper 保守が必要 |
| OTel 統合パスとの整合性 | △ otelconnect との二重計測リスク | ◎ 同一インターフェースで自然にチェーン | ○ Connect 側は自然、Skipper が2箇所に分散 |

## Notes

- Echo ミドルウェア単独では、RPC メソッド名が URI パスからの推測に依存し Connect の公式 API（`Spec().Procedure`）を経由しない。ストリーミング RPC / gRPC プロトコルでは HTTP ステータスが常に 200 になるため破綻する
- Connect インターセプタ単独では、非 RPC エンドポイント（管理画面 `/admin/*`）をカバーしない。product-mgmt だけ Echo ミドルウェアを追加する必要があり統一性が崩れる

## More Information

* [Connect RPC Interceptors ドキュメント](https://connectrpc.com/docs/go/interceptors/)
* [Echo Logger Middleware ドキュメント](https://echo.labstack.com/docs/middleware/logger)
* [otelconnect-go — Connect RPC 用 OpenTelemetry インターセプタ](https://github.com/connectrpc/otelconnect-go)
