---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# モノレポのパッケージマネージャー選定

## Decision Outcome

TypeScript パッケージ（React フロントエンド・Fastify BFF）の依存管理に pnpm を採用する。すべての評価軸で優れており、Vite 公式がスキャフォルド第一候補に採用し、Fastify モノレポのテンプレートでも事実上の標準となっている。symlink ベースの非 flat な node_modules 構造により phantom dependency を構造的に防止でき、マイクロサービス間の依存分離が明確になる。npm とコマンド体系がほぼ同一で移行・学習コストが最小限であり、4種のフィルタセレクタ（名前/glob/依存/Git差分）で選択的操作も柔軟に行える。

### Consequences

- symlink ベースの厳密な依存管理により、マイクロサービス間の依存分離が明確になる
- Vite + Fastify エコシステムとの親和性が高く、豊富なモノレポ構築事例を参考にできる
- ビルドキャッシュやタスクオーケストレーションが必要になった場合、Turborepo 等の追加導入を検討する必要がある（ただし Vite+ の monorepo タスクランナーが代替候補となる可能性がある）

## Context and Problem Statement

4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービスをモノレポで開発するにあたり、TypeScript パッケージ（React フロントエンド・Fastify BFF）の依存関係を効率的に管理できるパッケージマネージャーを選定する必要がある。Go サービス（ECサイト・商品管理・顧客管理）の依存管理は Go Modules で行うため、本 ADR のスコープは TypeScript パッケージに限定される。

## Prerequisites

- フロントエンド: React (Vite)、BFF: Fastify (TypeScript)、バックエンドAPI: Go (Echo) (ADR-0001)
- 4サービス（BFF・ECサイト・商品管理・顧客管理）をモノレポ構成で開発する（要件定義: docs/plans/requirements.md）
- 学習目的のプロジェクトである

## Decision Drivers

- ワークスペース管理の充実度
- 依存関係の分離・安全性
- React (Vite) + Fastify エコシステムとの相性
- 学習コストと情報の豊富さ

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **pnpm（採用）** | コンテンツアドレッサブルストアと symlink ベースの非 flat な node_modules 構造を特徴とするパッケージマネージャー |
| Yarn Berry (v4) | ワークスペース機能の元祖。PnP (Plug'n'Play) による node_modules 排除が特徴 |
| npm | Node.js に同梱されるデフォルトのパッケージマネージャー。追加インストール不要で互換性が最も高い |
| Bun | JavaScript ランタイム兼パッケージマネージャー。インストール速度が圧倒的に速い |

## Comparison Overview

| 観点 | pnpm | Yarn Berry (v4) | npm | Bun |
|------|------|-----------------|-----|-----|
| ワークスペース管理の充実度 | ◎ | ◎ | △ | ○ |
| 依存関係の分離・安全性 | ◎ | ○ | △ | ○ |
| Vite + Fastify との相性 | ◎ | △ | ○ | ○ |
| 学習コストと情報の豊富さ | ◎ | △ | ◎ | △ |
| インストール速度 | ○ | ○ | △ | ◎ |

## Notes

- Yarn Berry は PnP モードで Vite との互換性問題が複数報告されている（クラッシュ、ESM 解決エラー等）。node_modules モードに切り替えると PnP の優位性を失い、pnpm を選ぶ方が合理的
- Bun は Fastify をランタイムとして使う場合に OpenTelemetry 計装が未対応（oven-sh/bun#26536）。Node.js API 互換性が約95%でエッジケースの問題が残る
- pnpm のタスクキャッシュは持たず、必要なら Turborepo 等の追加が必要

## More Information

- pnpm 公式ドキュメント: https://pnpm.io
- pnpm ワークスペース: https://pnpm.io/workspaces
