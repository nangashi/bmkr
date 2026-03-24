---
status: "accepted"
date: 2026-03-23
last-validated: 2026-03-23
---

# ログ出力パターンに Echo ミドルウェア + Connect インターセプタの組み合わせを採用する

## Context and Problem Statement

log/slog を採用し（ADR-0015）、ログ設計ガイドで canonical log line パターンが定義済みだが、ログをどのレイヤーで出力するかが未決定である。Echo + Connect RPC 構成ではリクエストの流れが `Echo ミドルウェア → echo.WrapHandler → Connect インターセプタ → RPC ハンドラ` となり、各レイヤーでアクセスできる情報が異なる。この ADR ではログ出力の実装パターンを選定する。

## Prerequisites

* log/slog を採用済み (ADR-0015)
* バックエンド API は Go (Echo) を使用する (ADR-0001)
* サービス間通信に Connect RPC を採用している (ADR-0003)
* ログ設計ガイド (docs/guides/implementation/logging-strategy.md) で canonical log line パターンが定義済み — RPC メソッド名、ステータス(ok/error)、duration_ms、request_id を含む1行ログ
* product-mgmt は管理画面（templ + HTMX）も Echo ルートで提供している (ADR-0007)
* 3サービスで同一の実装パターンを使う（統一性）
* 将来の OpenTelemetry 統合を妨げない選択であること

## Decision Drivers

* canonical log line との適合性 — ガイドが求める「RPC メソッド名・ステータス・レイテンシ・request_id を含む1行ログ」をどこまで忠実に実現できるか
* ログカバレッジ — RPC エンドポイント・非 RPC エンドポイント（ヘルスチェック、管理画面）の両方をカバーできるか
* 実装・保守の複雑さ — 自作コード量、設定の複雑さ、二重出力等の罠を避けられるか
* OTel 統合パスとの整合性 — 将来 `otelconnect` インターセプタを追加する際に自然に共存できるか

## Considered Options

* Echo ミドルウェア
* Connect インターセプタ
* Echo ミドルウェア + Connect インターセプタの組み合わせ

### Excluded Options

* ハンドラ直接呼び出し — ログ設計ガイドで明確にアンチパターンとして定義。ハンドラ内での散発的なログ出力は canonical log line パターンに反し、ログの追加・削除がビジネスロジックに混在する

## Comparison Overview

| 判断軸 | Echo ミドルウェア | Connect インターセプタ | 組み合わせ |
|--------|-----------------|---------------------|-----------|
| canonical log line との適合性 | △ RPC メソッド名は URI パス推測、ストリーミング/gRPC で破綻 | ◎ 4フィールド全てネイティブ取得 | ◎ RPC は Connect で正確に、非 RPC は Echo でカバー |
| ログカバレッジ | ◎ RPC・非 RPC ともワンストップでカバー | △ 非 RPC（管理画面）が未カバー | ◎ 全エンドポイントをカバー、Skipper で二重出力防止 |
| 実装・保守の複雑さ | ◎ `RequestLoggerWithConfig` の設定だけ | ○ 30-40行の自作、管理画面は別途対応 | ○ 2レイヤー管理 + Skipper 保守が必要 |
| OTel 統合パスとの整合性 | △ otelconnect との二重計測リスク | ◎ 同一インターフェースで自然にチェーン | ○ Connect 側は自然、Skipper が2箇所に分散 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Echo ミドルウェア

* Good, because RPC・非 RPC（ヘルスチェック・管理画面）を `e.Use()` 一行で全カバーでき、ログカバレッジに穴がない
* Good, because Echo の `RequestLoggerWithConfig` で設定ベースに構築でき、自作コードが最小限
* Good, because `samber/slog-echo` を使えば OTel trace ID 連携やフィルタリングが簡単に実現可能
* Bad, because RPC メソッド名は URI パスからの推測に依存し、Connect の公式 API（`Spec().Procedure`）を経由しない
* Bad, because Connect エラーコード（`not_found` 等）を取得できず、HTTP ステータスコードのみ
* Bad, because ストリーミング RPC / gRPC プロトコルでは HTTP ステータスが常に 200 になるため破綻する

### Connect インターセプタ

* Good, because `Spec().Procedure` で RPC メソッド名、`CodeOf(err)` で Connect エラーコードをネイティブに取得でき、canonical log line の要件を完全に満たす
* Good, because `otelconnect.NewInterceptor()` と同一インターフェースでチェーンでき、OTel 統合が最も自然
* Good, because Echo フレームワークに非依存で、将来のルーター変更にも影響しない
* Bad, because 非 RPC エンドポイント（管理画面 `/admin/*`）をカバーしない。product-mgmt だけ Echo ミドルウェアを追加する必要があり統一性が崩れる
* Bad, because ビジネスフィールド（`customer_id` 等）の伝播に context 経由の共有バッファパターンが必要で、実装がやや複雑

### Echo ミドルウェア + Connect インターセプタの組み合わせ

* Good, because Connect インターセプタで RPC メタデータを正確に取得し、canonical log line の要件を完全に満たす
* Good, because Echo ミドルウェアで非 RPC エンドポイント（管理画面・ヘルスチェック）もカバーし、ログに穴がない
* Good, because 汎用 Skipper（パスに `.` を含むかで判定）を使えば、proto パッケージ追加時も Skipper 修正が不要
* Good, because OTel 導入時に Connect 側は otelconnect、Echo 側は otelecho と各レイヤーの標準ツールを使える
* Bad, because 2つのレイヤーを管理する必要があり、ログフォーマット変更時に両方の更新が必要
* Bad, because Skipper の設定保守が必要（汎用パターンで軽減可能だが、設計の複雑さは増す）
* Bad, because RPC ログと非 RPC ログでフォーマットが完全同一にはならない（RPC: `method: ec.v1.CartService/GetCart`、非 RPC: `method: GET /admin/products`）

## Decision Outcome

Chosen option: "Echo ミドルウェア + Connect インターセプタの組み合わせ"

### Rationale

ログカバレッジと canonical log line の適合性を重視した。Connect インターセプタで `Spec().Procedure` と `CodeOf(err)` を使い RPC メタデータを正確に取得しつつ、Echo ミドルウェアで非 RPC エンドポイント（管理画面・ヘルスチェック）もカバーする。Echo ミドルウェア単独ではストリーミング RPC / gRPC プロトコル対応時に破綻するリスクがあり、Connect インターセプタ単独では管理画面のログが欠けて3サービス統一が崩れる。組み合わせにより両方の弱点を補完できる。

### Accepted Tradeoffs

* 2レイヤー管理（Connect インターセプタ + Echo ミドルウェア）と Skipper 保守の複雑さを受け入れる。ログフォーマット変更時に両方の更新が必要になる
* RPC ログと非 RPC ログでフォーマットが完全同一にはならない（RPC: `method: ec.v1.CartService/GetCart`、非 RPC: `method: GET /admin/products`）

### Consequences

* Good, because RPC エンドポイントは Connect インターセプタで `Spec().Procedure` と `CodeOf(err)` を使い、canonical log line の4必須フィールドを正確に出力できる
* Good, because 非 RPC エンドポイント（管理画面・ヘルスチェック）は Echo ミドルウェアでカバーされ、ログに穴がない
* Good, because OTel 導入時に Connect 側は otelconnect、Echo 側は otelecho と各レイヤーの標準ツールを段階的に追加できる
* Bad, because RPC ログと非 RPC ログでフォーマットが完全同一にはならない

## More Information

* [Connect RPC Interceptors ドキュメント](https://connectrpc.com/docs/go/interceptors/)
* [Echo Logger Middleware ドキュメント](https://echo.labstack.com/docs/middleware/logger)
* [otelconnect-go — Connect RPC 用 OpenTelemetry インターセプタ](https://github.com/connectrpc/otelconnect-go)

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-23 | 初版作成 | N/A |
