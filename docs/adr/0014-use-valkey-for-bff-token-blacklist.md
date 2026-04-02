---
status: "accepted"
date: 2026-03-20
last-validated: 2026-03-20
---

# BFF の refresh token 失効管理用 KVS の選定

## Decision Outcome

Valkey を採用する。Redis 7.2.4 の完全な fork であり、ioredis からホスト名変更のみで利用できる API 互換性を持つ。BSD-3-Clause ライセンスにより学習成果物の公開に制約がなく、Redis のトリプルライセンスや DragonflyDB の BSL 1.1 と比較してシンプルである。Linux Foundation + AWS/Google/Oracle 等 50 社以上のバッキングにより、プロジェクトの継続性も十分と判断した。

### Accepted Tradeoffs

- 特になし。Redis に比べて日本語学習リソースが少ない点はあるが、Redis 互換のため Redis 向けの情報をそのまま適用でき、実質的な障壁にはならない

### Consequences

- ioredis のホスト名変更のみで BFF から利用でき、認証実装に集中できる
- BSD-3-Clause ライセンスにより、学習成果物を GitHub に公開する際のライセンス懸念がない
- 日本語の学習リソースは Redis に比べて限定的であり、トラブル時は Redis 向けの情報を読み替える必要がある

## Context and Problem Statement

BFF（Fastify/TypeScript）における refresh token の失効管理・ブラックリストの永続化先として KVS を選定する必要がある。BFF は RDB を持たないが、認証トークン状態管理用の KVS は許容されている（ADR-0004）。この ADR では KVS エンジンの選定のみを扱い、キー命名規則・TTL 値等のスキーマ設計や Go バックエンド側での KVS 利用は対象外とする。

## Prerequisites

- BFF は Fastify (TypeScript) で、RDB は持たないが認証トークン状態管理用 KVS は許容する (ADR-0001, ADR-0004)
- JWT 自前実装でトークンローテーションを行い、refresh token の失効管理が必要 (ADR-0005)
- 学習目的のプロジェクトであり、ローカル実行のみ（Docker Compose 前提）

## Decision Drivers

- Redis API 互換性
- ライセンスの明確さ
- プロジェクトの継続性
- ローカル開発環境でのセットアップ容易性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **Valkey（採用）** | Linux Foundation 主導の Redis 7.2.4 フォーク。BSD-3-Clause |
| Redis | オリジナルの in-memory データストア。AGPLv3/RSALv2/SSPLv1 トリプルライセンス |
| DragonflyDB | C++ で再設計された Redis 互換ストア。BSL 1.1 |
| Garnet | Microsoft Research 製の RESP 互換キャッシュストア。MIT |

除外: KeyDB（メインメンテナ離脱、2年以上リリースなし）、Kvrocks（SSD ベースで用途が異なる）、DiceDB（アーキテクチャ移行中で安定性に懸念）

## Comparison Overview

| 判断軸 | Valkey | Redis | DragonflyDB | Garnet |
|--------|--------|-------|-------------|--------|
| Redis API 互換性 | ◎ Redis 7.2.4 fork で完全互換。ioredis のホスト名変更のみで動作 | ◎ Redis 自身なので定義上 100% | ○ 基本コマンドは実機検証済み。一部高度なコマンド未サポート | △ RESP 互換で基本コマンドは動作。Node.js 検証事例が少ない |
| ライセンスの明確さ | ◎ BSD-3-Clause。制約なし | △ AGPLv3/RSALv2/SSPLv1 のトリプルライセンス。2年で2回変更歴 | △ BSL 1.1（非 OSS）。2029年に Apache 2.0 へ転換 | ◎ MIT。最も制約が少ない |
| プロジェクトの継続性 | ◎ Linux Foundation + 50社参画。antirez 参加 | ○ 高い ARR、antirez 復帰。外部コントリビューター激減 | ○ 活発な開発。BSL でフォーク制限 | ○ Microsoft Research + Azure 統合。外部採用事例が少ない |
| セットアップ容易性 | ◎ Alpine イメージ 3行で起動。Redis と同一ポート・設定 | ◎ Alpine イメージ 3行で起動。チュートリアル最豊富 | ○ 公式 docker-compose 提供。イメージサイズ大 | △ .NET ランタイム含みサイズ大。ulimits 設定必要 |

## Notes

- Valkey 8.x 以降、HELLO コマンドの `server` フィールドが `"valkey"` に変更されており、ioredis の古いバージョンで問題が発生した報告がある（最新版で修正済み）
- Redis は2年で2回のライセンス変更（BSD→RSAL/SSPL→RSAL/SSPL/AGPLv3）があり、ioredis がメンテナンスモード（best-effort）に移行している
- DragonflyDB の BSL 1.1 は公式に「Open Source ライセンスではない」と明示されている
- Garnet は Node.js クライアントとの組み合わせに特化した公式テスト・ドキュメントが存在せず、Stream 系コマンドと Functions が未実装

## More Information

* Valkey 公式: https://valkey.io/
* Valkey Docker Hub: https://hub.docker.com/r/valkey/valkey/
* Valkey Migration from Redis: https://valkey.io/topics/migration/
* iovalkey (Valkey 公式 Node.js クライアント): https://github.com/valkey-io/iovalkey
* ADR-0004 にて「BFF の認証トークン状態管理に必要な KVS は別途 ADR で決定する」と記載されており、本 ADR がその決定に該当する
