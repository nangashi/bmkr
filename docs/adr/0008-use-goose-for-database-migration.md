---
status: "accepted"
date: 2026-03-15
last-validated: 2026-03-15
---

# Go バックエンドの DB マイグレーションツールに goose を採用する

## Context and Problem Statement

ADR-0004 で PostgreSQL + sqlc を採用したが、sqlc にはマイグレーション機能が内蔵されておらず、スキーマ変更のバージョン管理・適用・ロールバックのために外部ツールが必要である。3つのバックエンドサービス（ECサイト・商品管理・顧客管理）がそれぞれ独立した DB を持つ Database-per-Service パターンのため、複数 DB のマイグレーションを管理できることも求められる。この ADR では Go バックエンドサービスの DB マイグレーション管理ツールおよび sqlc との連携方式を決定する。マイグレーションの運用ルール（命名規則、レビュープロセス等）や TypeScript 側の DB アクセスはスコープ外とする。

## Prerequisites

* バックエンドAPIは Go / Echo を使用 (ADR-0001)
* データベースに PostgreSQL、データアクセスライブラリに sqlc を採用済み (ADR-0004)
* ADR-0004 の Consequences に「goose 等の外部ツールが必須」と明記
* 各サービスが独立した DB を持つ Database-per-Service パターン (要件定義: docs/plans/requirements.md)
* 学習目的のプロジェクトである

## Decision Drivers

* sqlc ワークフローとの統合しやすさ — マイグレーション SQL を sqlc の schema として直接参照できるか、設定の複雑さ、公式ドキュメントの充実度
* マイグレーション失敗時の安全性 — 失敗時のロールバック挙動、dirty state からの復旧しやすさ、トランザクション内実行の対応
* 学習コストとドキュメントの充実度 — 初学者がキャッチアップするまでの時間、日本語・英語の学習リソースの量、概念のシンプルさ
* Database-per-Service 運用のしやすさ — 複数 DB に対するマイグレーション管理の方法、CI/CD での自動化のしやすさ、設定ファイルの管理方法

## Considered Options

* goose (pressly/goose)
* golang-migrate (golang-migrate/migrate)
* Atlas (ariga/atlas)
* dbmate (amacneil/dbmate)

### Excluded Options

* sql-migrate (rubenv/sql-migrate) — メンテナンス頻度が低く（最終 push 2025-11）、goose/golang-migrate と比較して差別化点が少ない
* tern (jackc/tern) — PostgreSQL 専用で pgx 作者が開発しているが、コミュニティが小さく（1.3K Stars）学習リソースが限定的
* pgroll (xataio/pgroll) — JSON 定義のため sqlc の公式パースに非対応。ゼロダウンタイムマイグレーションは学習プロジェクトではオーバースペック

## Comparison Overview

| 判断軸 | goose | golang-migrate | Atlas | dbmate |
|--------|-------|----------------|-------|--------|
| sqlc ワークフローとの統合しやすさ | ◎ 公式連携ガイドあり。マーカーはSQLコメントで干渉なし | ○ up/down分離でsqlcが自動認識。ソート順に注意 | ○ 公式統合ガイド2本あり。ただし手順が多い | ◎ タイムスタンプ命名で順序不一致なし。設定1行 |
| マイグレーション失敗時の安全性 | ◎ デフォルトTx実行。dirty state概念なし | △ dirty state問題あり。手動force必須 | ○ Tx+lint+sum。ただしlint/downはCE不可 | ○ デフォルトTx実行。dirty stateなし。ロック未実装 |
| 学習コストとドキュメントの充実度 | ◎ シンプル設計。日本語記事あり。設定ファイル不要 | ○ シンプル概念。18K Stars。公式ドキュメントは薄い | △ 宣言的パラダイムの学習コスト。CE/Pro区分が複雑 | ○ 極めてシンプル。日本語リソースはほぼ皆無 |
| Database-per-Service 運用のしやすさ | ○ Provider APIで独立管理可能。CLI切替も容易 | ○ フラグ切替で対応可能。一括管理機能はなし | ○ env+for_eachで管理可能。既知issueあり | ○ 環境変数+フラグ。create/dropコマンドが便利 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### goose (pressly/goose)

軽量でシンプルな SQL ベースのマイグレーションツール (10.3K Stars, MIT License)。SQL ファイルに `-- +goose Up` / `-- +goose Down` マーカーを記述する方式。Go 関数によるマイグレーションもサポート。

* Good, because マイグレーション SQL ファイルを sqlc の schema ディレクトリとして直接参照可能。公式連携ガイド（goose blog）が存在し、sqlc.yaml の設定例が具体的に示されている
* Good, because `-- +goose Up/Down` マーカーは SQL コメントなので sqlc のパースに干渉しない。sqlc は Up マイグレーションのみを解析し Down を無視する
* Good, because デフォルトで全マイグレーションがトランザクション内で実行される。PostgreSQL のトランザクショナル DDL を活用し、失敗時は自動ロールバック。golang-migrate のような dirty state が発生しない
* Good, because `-- +goose NO TRANSACTION` アノテーションで `CREATE INDEX CONCURRENTLY` 等のトランザクション外実行にも対応
* Good, because 概念が非常にシンプル。設定ファイル不要（No config files 設計思想）で、環境変数と CLI フラグだけで動作する
* Good, because 日本語リソースが複数存在（Qiita、CyberAgent Developers Blog、enish engineering blog 等）
* Good, because Provider API がグローバルステートを持たず、サービスごとに独立した Provider を生成可能。CLI でも `-dir` フラグでマイグレーションディレクトリを切り替えられる
* Good, because Go の `embed.FS` サポートでマイグレーション SQL をバイナリに埋め込み可能
* Bad, because シーケンシャル番号使用時はゼロパディング必須（sqlc との辞書順/数値順不整合を回避するため）。タイムスタンプ形式なら問題なし
* Bad, because `NO TRANSACTION` マイグレーション失敗時に dirty マークされず、部分適用リスクがある。べき等に記述する必要がある
* Bad, because dry-run（実行せずに適用される SQL を確認する）機能がない

### golang-migrate (golang-migrate/migrate)

Go マイグレーションツール中最多 Stars (18.2K Stars, MIT License)。CLI とライブラリの両方として利用可能。`{version}_{title}.up.sql` / `{version}_{title}.down.sql` のペアで管理。

* Good, because Go マイグレーションツール中最多 Stars で、StackOverflow やブログ記事でのカバー率が高い
* Good, because up.sql/down.sql 分離形式で sqlc が down を自動無視。sqlc 公式サポート
* Good, because `pg_advisory_lock` による排他制御で複数 Pod の同時実行を防止
* Good, because マイグレーションソースとして filesystem 以外に GitHub、S3、GCS 等をサポート
* Bad, because マイグレーション失敗・タイムアウト時に DB が dirty マークされ、手動 `force` コマンドが必須。メンテナーが意図的に wontfix としている
* Bad, because ファイル名のゼロパディング問題（sqlc との辞書順/数値順不整合）
* Bad, because 公式ドキュメントが最小限（GETTING_STARTED.md、MIGRATIONS.md、FAQ.md の3ファイル程度）
* Bad, because 最終リリースが 2025-11 でやや間隔が空いている。Open Issues が 458 件

### Atlas (ariga/atlas)

宣言的（declarative）とバージョン管理（versioned）の両アプローチをサポート (8.2K Stars)。HCL/SQL からスキーマを定義し、現在の DB 状態との diff を自動計算。

* Good, because 宣言的モードでは `schema.sql` を sqlc と共有でき、単一の真実の源として管理可能
* Good, because sqlc との統合ガイドが公式に2本存在（declarative/versioned）
* Good, because `atlas.sum` による整合性チェック、dev-database でのシミュレーション実行が可能
* Good, because GitHub Actions が14以上用意されており CI/CD 統合が充実
* Bad, because `migrate lint`（静的解析）と `migrate down`（ロールバック）が Community Edition では使用不可（Pro 必須）
* Bad, because Community Edition では views, triggers, functions 等の差分検出が不可
* Bad, because 宣言的パラダイムの学習コストが高い。HCL 設定言語、CE/Open/Pro の3層構造の理解が必要
* Bad, because versioned モードのワークフローが4ステップと手順が多い

### dbmate (amacneil/dbmate)

フレームワーク非依存の軽量マイグレーションツール (6.8K Stars, MIT License)。タイムスタンプベースのバージョニング。`schema.sql` ダンプ機能あり。

* Good, because タイムスタンプベースの命名で辞書順=時系列順が保証され、sqlc との順序不一致問題が発生しない
* Good, because デフォルトでトランザクション内実行。失敗時は自動ロールバック、dirty state にならない
* Good, because `dbmate create` / `dbmate drop` で DB 自体の作成・削除が可能。`dbmate wait` で DB 起動待機もできる
* Good, because `schema.sql` ダンプ機能でスキーマ全体像を常に確認・git diff 可能
* Bad, because 日本語リソースがほぼ皆無（Zenn 検索結果 0 件）
* Bad, because 並行マイグレーション実行時のアドバイザリロック未実装
* Bad, because `lib/pq` ドライバ使用で、`pgx` への移行が未完了

## Decision Outcome

Chosen option: "goose (pressly/goose)", because 広く実績があり機能が十分であるため。

### Rationale

goose は4つの判断軸すべてにおいて安定した評価を示した。sqlc との統合では公式連携ガイドが存在し設定がシンプル、マイグレーション失敗時の安全性ではデフォルトのトランザクション実行と dirty state が発生しない設計、学習コストでは設定ファイル不要のシンプルさと日本語リソースの存在、Database-per-Service 運用では Provider API による独立管理が可能という点で、特定の軸に突出するわけではないが全体的にバランスが取れている。10.3K Stars の実績とアクティブなメンテナンス状況も安心材料である。

### Accepted Tradeoffs

特になし。dry-run 機能の欠如や NO TRANSACTION 時の部分適用リスクは存在するが、学習プロジェクトにおいては実用上の問題にならないと判断した。

### Consequences

* Good, because sqlc のマイグレーションディレクトリをそのまま schema として参照でき、スキーマ定義の一元管理が実現する
* Good, because 設定ファイル不要のシンプルな設計により、3サービス分のマイグレーション管理を低い導入コストで開始できる
* Bad, because dry-run 機能がないため、本番適用前の確認は手動での SQL レビューに依存する

## More Information

* goose 公式ドキュメント: https://pressly.github.io/goose/
* goose + sqlc 連携ガイド: https://pressly.github.io/goose/blog/2024/goose-sqlc/
* sqlc DDL ドキュメント: https://docs.sqlc.dev/en/stable/howto/ddl.html
* goose GitHub: https://github.com/pressly/goose
