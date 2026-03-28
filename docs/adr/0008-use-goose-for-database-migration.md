---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# DB マイグレーションツールの選定

## Decision Outcome

goose (pressly/goose) を採用する。

goose は4つの判断軸すべてにおいて安定した評価を示した。sqlc との統合では公式連携ガイドが存在し設定がシンプル、マイグレーション失敗時はデフォルトのトランザクション実行と dirty state が発生しない設計で安全、学習コストでは設定ファイル不要のシンプルさと日本語リソースの存在、Database-per-Service 運用では Provider API による独立管理が可能という点で、全体的にバランスが取れている。

### Accepted Tradeoffs

- dry-run 機能の欠如や NO TRANSACTION 時の部分適用リスクは存在するが、学習プロジェクトにおいては実用上の問題にならないと判断した

### Consequences

- sqlc のマイグレーションディレクトリをそのまま schema として参照でき、スキーマ定義の一元管理が実現する
- 設定ファイル不要のシンプルな設計により、3サービス分のマイグレーション管理を低い導入コストで開始できる
- dry-run 機能がないため、本番適用前の確認は手動での SQL レビューに依存する

## Context and Problem Statement

ADR-0004 で PostgreSQL + sqlc を採用したが、sqlc にはマイグレーション機能が内蔵されておらず、スキーマ変更のバージョン管理・適用・ロールバックのために外部ツールが必要である。3つのバックエンドサービス（ECサイト・商品管理・顧客管理）がそれぞれ独立した DB を持つ Database-per-Service パターンのため、複数 DB のマイグレーションを管理できることも求められる。マイグレーションの運用ルール（命名規則、レビュープロセス等）や TypeScript 側の DB アクセスはスコープ外とする。

## Prerequisites

- バックエンド: Go (Echo) (ADR-0001)、DB: PostgreSQL + sqlc (ADR-0004)
- ADR-0004 の Consequences に「goose 等の外部ツールが必須」と明記
- Database-per-Service パターン (要件定義)、学習目的

## Decision Drivers

- sqlc ワークフローとの統合しやすさ
- マイグレーション失敗時の安全性
- 学習コストとドキュメントの充実度
- Database-per-Service 運用のしやすさ

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **goose（採用）** | SQL ベースの軽量マイグレーションツール。`-- +goose Up/Down` マーカー方式 |
| golang-migrate | CLI とライブラリ両用の Go マイグレーションツール。up/down ファイル分離方式 |
| Atlas | 宣言的とバージョン管理の両アプローチをサポート。HCL/SQL 定義から diff を自動計算 |
| dbmate | フレームワーク非依存の軽量ツール。タイムスタンプベースのバージョニング |

除外: sql-migrate（メンテナンス頻度が低い）、tern（コミュニティが小さく学習リソースが限定的）、pgroll（JSON 定義のため sqlc パース非対応、ゼロダウンタイムは過剰）

## Comparison Overview

| 判断軸 | goose | golang-migrate | Atlas | dbmate |
|--------|-------|----------------|-------|--------|
| sqlc ワークフローとの統合しやすさ | ◎ 公式連携ガイドあり。マーカーはSQLコメントで干渉なし | ○ up/down分離でsqlcが自動認識。ソート順に注意 | ○ 公式統合ガイド2本あり。ただし手順が多い | ◎ タイムスタンプ命名で順序不一致なし。設定1行 |
| マイグレーション失敗時の安全性 | ◎ デフォルトTx実行。dirty state概念なし | △ dirty state問題あり。手動force必須 | ○ Tx+lint+sum。ただしlint/downはCE不可 | ○ デフォルトTx実行。dirty stateなし。ロック未実装 |
| 学習コストとドキュメントの充実度 | ◎ シンプル設計。日本語記事あり。設定ファイル不要 | ○ シンプル概念。公式ドキュメントは薄い | △ 宣言的パラダイムの学習コスト。CE/Pro区分が複雑 | ○ 極めてシンプル。日本語リソースはほぼ皆無 |
| Database-per-Service 運用のしやすさ | ○ Provider APIで独立管理可能。CLI切替も容易 | ○ フラグ切替で対応可能。一括管理機能はなし | ○ env+for_eachで管理可能。既知issueあり | ○ 環境変数+フラグ。create/dropコマンドが便利 |

## Notes

- **golang-migrate の dirty state 問題**: マイグレーション失敗・タイムアウト時に DB が dirty マークされ、手動 `force` コマンドが必須。メンテナーが意図的に wontfix としている
- **Atlas CE の機能制限**: `migrate lint`（静的解析）と `migrate down`（ロールバック）が Community Edition では使用不可（Pro 必須）。views, triggers, functions 等の差分検出も CE では不可
- **dbmate の並行実行リスク**: アドバイザリロック未実装のため、複数 Pod の同時実行を防止できない
- **goose の NO TRANSACTION 注意**: `-- +goose NO TRANSACTION` 使用時は失敗しても dirty マークされず、部分適用リスクがあるため冪等に記述する必要がある

## More Information

- goose 公式ドキュメント: https://pressly.github.io/goose/
- goose + sqlc 連携ガイド: https://pressly.github.io/goose/blog/2024/goose-sqlc/
- sqlc DDL ドキュメント: https://docs.sqlc.dev/en/stable/howto/ddl.html
- goose GitHub: https://github.com/pressly/goose
