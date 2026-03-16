---
status: "accepted"
date: 2026-03-15
---

# データベースに PostgreSQL、データアクセスライブラリに sqlc を採用する

## Context and Problem Statement

学習用模擬ECサイトの3つのバックエンドAPI（ECサイト・商品管理・顧客管理）がそれぞれ独立したDBを持つ。データベースエンジンおよびGoのデータアクセスライブラリ（ORM・クエリビルダ）を選定する必要がある。BFFはAPIゲートウェイでありDBを持たないため、TypeScript側のORM選定は本ADRのスコープ外とし、管理画面UIの技術選定時に必要に応じて別途決定する。

## Prerequisites

* バックエンドAPI（ECサイト・商品管理・顧客管理）は Go / Echo を使用する (ADR-0001)
* BFF は TypeScript / Fastify でDBを持たない (ADR-0001)
* 各サービスが独立したDBを持つ Database-per-Service パターン (要件定義: docs/plans/requirements.md)
* サービス間通信プロトコルは ADR-0003 で決定済み
* 学習目的のプロジェクトである（商用運用ではない）

## Decision Drivers

* SQL・RDB設計の学習効果 — SQLクエリ、インデックス設計、正規化などRDBの基礎をどれだけ深く学べるか
* Go言語イディオムとの親和性 — Goの慣習（struct, interface, error handling, コード生成パターン）に沿った設計か
* マイグレーション・スキーマ管理の充実度 — スキーマ変更のバージョン管理、適用・ロールバックの仕組みが整っているか
* Echo との統合実績と学習リソース — Echo フレームワークとの組み合わせ事例、チュートリアル・サンプルコードの豊富さ

## Considered Options

* PostgreSQL + sqlc
* PostgreSQL + GORM
* PostgreSQL + Ent
* MySQL + GORM

## Pros and Cons of the Options

### PostgreSQL + sqlc

SQLクエリファイルからタイプセーフなGoコードを自動生成するコードジェネレーター (17.1K Stars)。ORM ではなく、手書きSQLのパフォーマンスを維持しつつボイラープレートを削減する。

* Good, because SQL直書きが必須のため、SELECT/JOIN/サブクエリ・DDL（CREATE TABLE/ALTER TABLE）などRDBの基礎を実践的に学べる
* Good, because コンパイル時にSQLの文法・型の整合性を検証し、フィードバックループが短い
* Good, because 生成コードが `context.Context` 第一引数・`DBTX` interface・`error` 返却などGoイディオムに完全準拠。`go generate` パイプラインもGo標準パターン
* Good, because `emit_interface: true` で生成される `Querier` interfaceにより、テスト時のモック差し替えが容易。Goの「小さなinterfaceで依存性を注入する」イディオムに合致
* Good, because pgx/v5ネイティブサポートでPostgreSQLの高度な機能（LISTEN/NOTIFY, COPYプロトコル等）も利用可能
* Good, because パフォーマンスはdatabase/sqlとほぼ同等（GORM比で約2倍高速: 15,000行取得時 ~31.7ms vs ~59.3ms）
* Bad, because マイグレーション機能を内蔵せず、goose/golang-migrate等の外部ツールが必須（ツールチェーンが2つになる）
* Bad, because 動的クエリ（実行時の条件分岐WHERE）のサポートが弱く、SQL側でCASE/COALESCEを使うか複数クエリを定義する必要がある
* Bad, because 「sqlc + Echo」に特化したチュートリアルは限定的。サンプルリポジトリは存在するが一気通貫の教材はない

### PostgreSQL + GORM

Goで最も人気のあるフル機能ORM (39.6K Stars)。structタグベースのコードファーストアプローチで、AutoMigrate、Hooks、Preload等を備える。

* Good, because 「Echo + GORM + PostgreSQL」はGo Web開発の定番構成で、Medium・DEV Community・Qiita等に多数のチュートリアルが存在
* Good, because AutoMigrateでstruct定義からテーブルを自動生成でき、開発初期の立ち上げが非常に速い
* Good, because Hooks（BeforeCreate等）・Preload・AssociationなどECサイトのリレーション多めのドメインと相性がよい
* Bad, because メソッドチェーンでSQLが抽象化され、SELECT/JOIN/サブクエリを自分で書く機会が大幅に減る
* Bad, because リフレクション多用・暗黙的挙動（デフォルトsoft delete、自動タイムスタンプ等）がGoの「明示的であること」の哲学に反する
* Bad, because AutoMigrateにはロールバック・バージョン管理がなく、本格運用には外部ツールが必要

### PostgreSQL + Ent

Meta（Facebook）発のコードファーストORM (16.9K Stars)。Goコードでスキーマを定義し、タイプセーフなAPIをコード生成する。

* Good, because Atlas連携によるバージョン管理マイグレーションが最も充実（lint による破壊的変更検出、ロールバック、進捗追跡）
* Good, because コード生成ベースで静的型付けされたAPIを提供し、コンパイル時に型安全性を保証
* Good, because `go generate` によるコード生成はGoエコシステムの標準パターンに合致
* Bad, because Ent独自のDSL（`field.Int("age").Positive()`, `edge.To("pets", Pet.Type)`）は他のGoプロジェクトに転用しにくい
* Bad, because 「Echo + Ent」特化のチュートリアルは少なく、日本語リソースもかなり限定的
* Bad, because まだv0.14.0でv1.0未到達。コミュニティサポートへの不満も報告されている

### MySQL + GORM

PostgreSQLではなくMySQLを採用し、Go最大のORM であるGORMと組み合わせる構成。

* Good, because MySQLは日本語の学習リソースが最も豊富。日本国内のフリーランス求人でもMySQL需要が高い
* Good, because PostgreSQLとは異なるRDBMSを経験でき（InnoDBクラスタインデックス、レプリケーション方式等）、比較学習の価値がある
* Bad, because SQL標準への準拠度がPostgreSQLより低い（FULL OUTER JOIN未サポート等）
* Bad, because JSONB型・配列型・Range型などPostgreSQLの高度なデータ型が存在せず、柔軟性に劣る
* Bad, because GORMのリフレクション・暗黙的挙動の問題はPostgreSQL版と同様

## Decision Outcome

Chosen option: "PostgreSQL + sqlc", because Go言語イディオムとの親和性を重視するため。生成コードが context.Context・interface・error handling などGoの慣習に完全準拠し、`go generate` パイプラインもGo標準パターンに合致する。また、SQL直書きによりRDB設計の基礎も実践的に学べる。

### Consequences

* Good, because SQLを直接書く開発スタイルにより、RDB設計とGoイディオムの両方を実践的に習得できる
* Good, because pgx/v5ネイティブサポートによりPostgreSQLの高度な機能も将来的に活用可能
* Bad, because マイグレーション管理にgoose等の外部ツールを別途導入する必要があり、ツールチェーンが増える
* Bad, because 動的クエリの構築が難しく、検索条件の組み合わせが複雑な画面ではSQL側での工夫（CASE/COALESCE）または複数クエリの定義が必要

## More Information

* sqlc 公式ドキュメント: https://docs.sqlc.dev/
* PostgreSQL Getting Started (sqlc): https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html
* goose（マイグレーションツール候補）: https://github.com/pressly/goose
* sqlc + goose 連携ガイド: https://pressly.github.io/goose/blog/2024/goose-sqlc/
* TypeScript側のORM/クエリビルダは、管理画面UIの技術選定時に必要に応じて別途決定する
