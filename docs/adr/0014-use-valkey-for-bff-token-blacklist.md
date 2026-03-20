---
status: "accepted"
date: 2026-03-20
last-validated: 2026-03-20
---

# BFF の refresh token 失効管理用 KVS に Valkey を採用する

## Context and Problem Statement

BFF（Fastify/TypeScript）における refresh token の失効管理・ブラックリストの永続化先として KVS を選定する必要がある。BFF は RDB を持たないが、認証トークン状態管理用の KVS は許容されている（ADR-0004）。この ADR では KVS エンジンの選定のみを扱い、キー命名規則・TTL 値等のスキーマ設計や Go バックエンド側での KVS 利用は対象外とする。

## Prerequisites

* BFF は Fastify (TypeScript) を使用する (ADR-0001)
* BFF は RDB を持たないが、認証トークン状態管理用の KVS は許容する (ADR-0004)
* JWT 自前実装でトークンローテーションを行い、refresh token の失効管理が必要 (ADR-0005)
* 学習目的のプロジェクトであり、ローカル実行のみ（Docker Compose 前提）

## Decision Drivers

* Redis API 互換性 — ioredis/node-redis クライアントでそのまま利用できるか、互換性の問題や制限がないか
* ライセンスの明確さ — 学習プロジェクトの参考実装として利用・公開する際のライセンス形態の透明性と制約
* プロジェクトの継続性 — メンテナンスが継続される見込み、リリース頻度、コミュニティの活発さ、バッキング組織
* ローカル開発環境でのセットアップ容易性 — Docker Compose での起動の簡潔さ、公式 Docker イメージの品質、設定の少なさ

## Considered Options

* Valkey
* Redis
* DragonflyDB
* Garnet

### Excluded Options

* KeyDB — メインメンテナが 2025-01 に離脱し、2年以上リリースなし。メンテナ自身が Valkey への移行を推奨
* Kvrocks — SSD ベースのストレージ（RocksDB）で大規模データセット向け。refresh token ブラックリストのようなインメモリ用途には過剰
* DiceDB — Valkey ベースにアーキテクチャを大幅変更中（Go → C への移行）。安定性に懸念

## Comparison Overview

| 判断軸 | Valkey | Redis | DragonflyDB | Garnet |
|--------|--------|-------|-------------|--------|
| Redis API 互換性 | ◎ Redis 7.2.4 fork で完全互換。ioredis のホスト名変更のみで動作 | ◎ Redis 自身なので定義上 100% | ○ 基本コマンドは実機検証済み。一部高度なコマンド未サポート | △ RESP 互換で基本コマンドは動作。Node.js 検証事例が少ない |
| ライセンスの明確さ | ◎ BSD-3-Clause。制約なし | △ AGPLv3/RSALv2/SSPLv1 のトリプルライセンス。2年で2回変更歴 | △ BSL 1.1（非 OSS）。2029年に Apache 2.0 へ転換 | ◎ MIT。最も制約が少ない |
| プロジェクトの継続性 | ◎ Linux Foundation + 50社参画。antirez 参加。月間 PR 80件 | ○ $300M ARR、antirez 復帰。外部コントリビューター激減 | ○ 毎月 100+ commits。BSL でフォーク制限 | ○ Microsoft Research + Azure 統合。外部採用事例が少ない |
| セットアップ容易性 | ◎ Alpine イメージ 3行で起動。Redis と同一ポート・設定 | ◎ Alpine イメージ 3行で起動。チュートリアル最豊富 | ○ 公式 docker-compose 提供。イメージ 190MB | △ .NET ランタイム含みサイズ大。ulimits 設定必要 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Valkey

Linux Foundation 主導の Redis 7.2.4 フォーク（25.2K Stars, BSD-3-Clause）。AWS ElastiCache / Google Memorystore のデフォルト。最新 v9.0.3（2026-02）。

* Good, because Redis 7.2.4 fork で RESP2/RESP3・コマンドセット・ポート・設定形式が完全互換。ioredis のホスト名変更のみで動作する
* Good, because BSD-3-Clause ライセンスで、学習成果物を GitHub に公開する際のライセンス懸念がゼロ
* Good, because Linux Foundation がホストし、AWS/Google/Oracle/Ericsson 等 50 社以上が参画。antirez（Redis 原作者）も最多コントリビューター（7,037 commits）として参加
* Good, because Docker Hub 公式イメージ `valkey/valkey:9-alpine` が提供されており、3行の docker-compose で起動可能
* Good, because AWS ElastiCache / Google Memorystore のデフォルトとなっており、学習した知識がクラウド環境でも活きる
* Bad, because Valkey 8.x 以降、HELLO コマンドの `server` フィールドが `"valkey"` に変更されており、ioredis の古いバージョンで問題が発生した報告がある（最新版で修正済み）
* Bad, because 「Valkey」自体の認知度は Redis に比べてまだ低く、日本語の学習リソースは Redis に比べて限定的

### Redis

オリジナルの in-memory データストア（73.5K Stars, AGPLv3/RSALv2/SSPLv1 トリプルライセンス）。最新 v8.6.1（2026-02）。

* Good, because KVS のデファクトスタンダードであり、学習した知識が実務でそのまま活きる。学習リソースが圧倒的に豊富
* Good, because `@fastify/redis` プラグインが公式に存在し、Fastify のデコレータパターンでシームレスに統合可能
* Good, because Redis 8 で AGPLv3（OSI 承認）が選択肢に追加され、OSS ライセンスとして利用可能に
* Bad, because 2年で2回のライセンス変更（BSD→RSAL/SSPL→RSAL/SSPL/AGPLv3）があり、ライセンスの安定性に懸念
* Bad, because ライセンス変更後、外部コントリビューターの大半を失い、コミュニティの主力が Valkey に移行
* Bad, because ioredis がメンテナンスモード（best-effort）に移行し、Redis Inc. は node-redis を推奨

### DragonflyDB

C++ で再設計された Redis 互換ストア（30.2K Stars, BSL 1.1）。Redis の最大 25 倍のスループット。最新 v1.37.0（2026-02）。

* Good, because マルチスレッドで Redis の最大 25 倍のスループット、80% のメモリ削減。アイドル時メモリ消費 21MB と軽量
* Good, because refresh token ブラックリストに必要な全コマンド（SET/EX, GET, DEL, EXISTS）が ioredis で実機検証済み
* Good, because 毎月 100+ commits と開発が極めて活発
* Bad, because BSL 1.1 は公式に「Open Source ライセンスではない」と明示されている
* Bad, because 一部高度なコマンド（WAIT, WAITAOF, FUNCTION 系等）が未サポート。エラーメッセージの型が Redis と異なる場合がある
* Bad, because Redis 固有概念（クラスタリング、Sentinel、RDB/AOF 等）の学習にはならない

### Garnet

Microsoft Research 製の RESP プロトコル互換キャッシュストア（11.8K Stars, MIT）。.NET ランタイム上で動作。最新 v1.1.1（2026-03）。

* Good, because MIT ライセンスで、4選択肢中最も制約が少ない
* Good, because Microsoft Research が開発し、Azure Cosmos DB に統合。週1回以上のリリースペース
* Bad, because Node.js クライアント（ioredis/node-redis）との組み合わせに特化した公式テスト・ドキュメントが存在しない
* Bad, because Stream 系コマンドと Functions が未実装。MSET が非アトミック
* Bad, because .NET ランタイム依存で Docker イメージサイズが大きく、ulimits 設定が必要。学習リソースが極めて限定的

## Decision Outcome

Chosen option: "Valkey", because Redis 互換でライセンスが明確（BSD-3-Clause）であり、Linux Foundation 主導で今後の利用拡大が見込まれるため。

### Rationale

Redis API 互換性とライセンスの明確さを重視した。Valkey は Redis 7.2.4 の完全な fork であり、ioredis からホスト名変更のみで利用できる API 互換性を持つ。BSD-3-Clause ライセンスにより学習成果物の公開に制約がなく、Redis のトリプルライセンスや DragonflyDB の BSL 1.1 と比較してシンプルである。Linux Foundation + AWS/Google/Oracle 等 50 社以上のバッキングにより、プロジェクトの継続性も十分と判断した。

### Accepted Tradeoffs

* 特になし。Redis に比べて日本語学習リソースが少ない点はあるが、Redis 互換のため Redis 向けの情報をそのまま適用でき、実質的な障壁にはならない

### Consequences

* Good, because ioredis のホスト名変更のみで BFF から利用でき、認証実装に集中できる
* Good, because BSD-3-Clause ライセンスにより、学習成果物を GitHub に公開する際のライセンス懸念がない
* Bad, because 日本語の学習リソースは Redis に比べて限定的であり、トラブル時は Redis 向けの情報を読み替える必要がある

## More Information

* Valkey 公式: https://valkey.io/
* Valkey Docker Hub: https://hub.docker.com/r/valkey/valkey/
* Valkey Migration from Redis: https://valkey.io/topics/migration/
* iovalkey (Valkey 公式 Node.js クライアント): https://github.com/valkey-io/iovalkey
* ADR-0004 にて「BFF の認証トークン状態管理に必要な KVS は別途 ADR で決定する」と記載されており、本 ADR がその決定に該当する

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-20 | 初版作成 | N/A |
