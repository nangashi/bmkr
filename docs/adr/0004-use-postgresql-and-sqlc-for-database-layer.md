---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# データベースエンジンとGoデータアクセスライブラリの選定

## Decision Outcome

データベースに PostgreSQL、データアクセスライブラリに sqlc を採用する。Go 言語イディオムとの親和性を最も重視した選定である。sqlc の生成コードは context.Context 第一引数・DBTX interface・error 返却など Go の慣習に完全準拠し、`go generate` パイプラインも Go 標準パターンに合致する。SQL 直書きが必須のため SELECT/JOIN/サブクエリ・DDL など RDB の基礎を実践的に学べる。コンパイル時に SQL の文法・型の整合性を検証しフィードバックループが短い。`emit_interface: true` で生成される Querier interface により、テスト時のモック差し替えも容易である。

### Consequences

- SQL を直接書く開発スタイルにより、RDB 設計と Go イディオムの両方を実践的に習得できる
- pgx/v5 ネイティブサポートにより PostgreSQL の高度な機能も将来的に活用可能
- マイグレーション管理に goose 等の外部ツールを別途導入する必要があり、ツールチェーンが増える
- 動的クエリの構築が難しく、検索条件の組み合わせが複雑な画面では SQL 側での工夫（CASE/COALESCE）または複数クエリの定義が必要

## Context and Problem Statement

学習用模擬ECサイトの3つのバックエンドAPI（ECサイト・商品管理・顧客管理）がそれぞれ独立したDBを持つ。データベースエンジンおよびGoのデータアクセスライブラリ（ORM・クエリビルダ）を選定する必要がある。BFFはAPIゲートウェイでありRDB（PostgreSQL）を持たないため、TypeScript側のORM選定は本ADRのスコープ外とし、管理画面UIの技術選定時に必要に応じて別途決定する。なお、BFF の認証トークン状態管理に必要な KVS（Redis互換ストア等）は本 ADR のスコープ外であり、別途 ADR で決定する。

## Prerequisites

- バックエンドAPI: Go / Echo (ADR-0001)。BFF: TypeScript / Fastify で RDB を持たない (ADR-0001)
- 各サービスが独立した DB を持つ Database-per-Service パターン（要件定義: docs/plans/requirements.md）
- サービス間通信プロトコルは ADR-0003 で決定済み
- 学習目的のプロジェクトである（商用運用ではない）

## Decision Drivers

- SQL・RDB設計の学習効果
- Go言語イディオムとの親和性
- マイグレーション・スキーマ管理の充実度
- Echo との統合実績と学習リソース

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **PostgreSQL + sqlc（採用）** | SQL クエリファイルからタイプセーフな Go コードを自動生成するコードジェネレーター。手書き SQL のパフォーマンスを維持しつつボイラープレートを削減 |
| PostgreSQL + GORM | Go で最も人気のあるフル機能 ORM。struct タグベースのコードファーストアプローチ |
| PostgreSQL + Ent | Meta（Facebook）発のコードファースト ORM。Go コードでスキーマを定義し、タイプセーフな API をコード生成 |
| MySQL + GORM | MySQL を採用し、Go 最大の ORM である GORM と組み合わせる構成 |

## Comparison Overview

| 観点 | PostgreSQL + sqlc | PostgreSQL + GORM | PostgreSQL + Ent | MySQL + GORM |
|------|-------------------|-------------------|------------------|--------------|
| SQL・RDB設計の学習効果 | ◎ | △ | △ | △ |
| Go言語イディオムとの親和性 | ◎ | △ | ○ | △ |
| マイグレーション・スキーマ管理 | △ | △ | ◎ | △ |
| Echo との統合実績と学習リソース | ○ | ◎ | △ | ○ |
| 型安全性 | ◎ | △ | ◎ | △ |

## Notes

- GORM はメソッドチェーンで SQL が抽象化され、SELECT/JOIN/サブクエリを自分で書く機会が大幅に減る。リフレクション多用・暗黙的挙動（デフォルト soft delete、自動タイムスタンプ等）が Go の「明示的であること」の哲学に反する
- Ent はまだ v0.14.0 で v1.0 未到達。独自 DSL は他の Go プロジェクトに転用しにくい
- sqlc は動的クエリのサポートが弱く、SQL 側で CASE/COALESCE を使うか複数クエリを定義する必要がある
- sqlc + Echo に特化したチュートリアルは限定的で、サンプルリポジトリは存在するが一気通貫の教材はない
- MySQL は SQL 標準への準拠度が PostgreSQL より低い（FULL OUTER JOIN 未サポート等）

## More Information

- sqlc 公式ドキュメント: https://docs.sqlc.dev/
- PostgreSQL Getting Started (sqlc): https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html
- goose（マイグレーションツール候補）: https://github.com/pressly/goose
- sqlc + goose 連携ガイド: https://pressly.github.io/goose/blog/2024/goose-sqlc/
- TypeScript側のORM/クエリビルダは、管理画面UIの技術選定時に必要に応じて別途決定する
- BFF の認証トークン状態管理（refresh token 失効・ブラックリスト等）に必要な KVS の選定は、本 ADR のスコープ外として別途 ADR で決定する
