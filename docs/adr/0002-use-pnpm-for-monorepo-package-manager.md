---
status: "implemented"
date: 2026-03-15
last-validated: 2026-03-15
---

# モノレポのパッケージマネージャーに pnpm を採用する

## Context and Problem Statement

4サービス（BFF・ECサイト・商品管理・顧客管理）のマイクロサービスをモノレポで開発するにあたり、TypeScript パッケージ（React フロントエンド・Fastify BFF）の依存関係を効率的に管理できるパッケージマネージャーを選定する必要がある。Go サービス（ECサイト・商品管理・顧客管理）の依存管理は Go Modules で行うため、本 ADR のスコープは TypeScript パッケージに限定される。

## Prerequisites

* ECサイトのフロントエンドは React (Vite)、BFF は Fastify (TypeScript)、バックエンドAPIは Go (Echo) を使用する (ADR-0001)
* 4サービス（BFF・ECサイト・商品管理・顧客管理）をモノレポ構成で開発する（要件定義: docs/plans/requirements.md）
* 学習目的のプロジェクトである

## Decision Drivers

* ワークスペース管理の充実度 — サービスごとの選択的インストール・ビルド・フィルタリング等、モノレポ運用に必要な機能がどれだけ揃っているか
* 依存関係の分離・安全性 — サービス間の phantom dependency 防止や依存の厳密さ。マイクロサービスごとの独立性を保てるか
* React (Vite) + Fastify エコシステムとの相性 — Vite / Fastify プロジェクトでの実績、既知の互換性問題の有無
* 学習コストと情報の豊富さ — セットアップの容易さ、npm との差分の大きさ、モノレポ構成に関するチュートリアル・記事の充実度

## Considered Options

* pnpm
* Yarn Berry (v4)
* npm
* Bun

## Pros and Cons of the Options

### pnpm

コンテンツアドレッサブルストアと symlink ベースの非 flat な node_modules 構造を特徴とするパッケージマネージャー。Vite エコシステムで事実上の標準として広く採用されている。

* Good, because Vite 公式がスキャフォルド第一候補に採用し、Fastify モノレポのテンプレートも pnpm workspace が事実上の標準
* Good, because 4種のフィルタセレクタ（名前/glob/依存/Git差分）で選択的操作が非常に柔軟
* Good, because symlink ベースの非 flat な node_modules で phantom dependency を構造的に防止
* Good, because npm とコマンド体系がほぼ同一で移行・学習コストが最小限
* Good, because クリーンインストールが npm 比約4倍高速
* Good, because コンテンツアドレッサブルストアによりディスク使用量を大幅削減
* Bad, because タスクキャッシュは持たず、必要なら Turborepo 等の追加が必要（ただし Vercel 非依存のため Turborepo の優位性は低下）
* Bad, because strict な node_modules 構造により一部古いパッケージで設定調整が必要な場合がある

### Yarn Berry (v4)

ワークスペース機能の元祖。PnP (Plug'n'Play) による node_modules 排除が革新的な特徴。

* Good, because `foreach --since` による変更検知ビルド、`focus --production` による最小インストール等、モノレポ専用機能が豊富
* Good, because constraints によるワークスペース間のルール強制が可能
* Bad, because PnP モードで Vite との互換性問題が複数報告されている（クラッシュ、ESM 解決エラー等、Vite issue #15910, #4307）
* Bad, because node_modules モードに切り替えると PnP の依存分離の優位性を失い、pnpm を選ぶ方が合理的
* Bad, because PnP という独自概念の学習負荷が高く、2025-2026年の最新情報も相対的に少ない

### npm

Node.js に同梱されるデフォルトのパッケージマネージャー。追加インストール不要で互換性が最も高い。

* Good, because Node.js 同梱で追加インストール不要、学習コストがゼロに近い
* Good, because 情報量が最も多く、トラブルシューティングが容易
* Bad, because ホイスティングによる phantom dependency 問題があり、サービス間の依存分離ができない
* Bad, because ワークスペースの選択的操作・ビルド順序制御が限定的
* Bad, because `workspace:*` プロトコル非対応でワークスペース間参照の管理が弱い
* Bad, because インストール速度が他の選択肢より大幅に遅い

### Bun

JavaScript ランタイム兼パッケージマネージャー。インストール速度が圧倒的に速い。

* Good, because インストール速度が圧倒的に速い（npm 比最大22倍）
* Good, because Next.js 非依存となったことで、パッケージマネージャーとしての互換性リスクが軽減された
* Good, because `--linker=isolated` で pnpm 同等の厳密な依存分離が可能
* Good, because `catalog:` による依存バージョン一元管理が可能
* Bad, because Fastify をランタイムとして使う場合、OpenTelemetry 計装が未対応（oven-sh/bun#26536）
* Bad, because モノレポ固有のトラブルシューティング情報が pnpm より少ない
* Bad, because Node.js API 互換性が約95%で、エッジケースで問題が発生する可能性がある

## Decision Outcome

Chosen option: "pnpm", because すべての評価軸で優れており、学習対象としても適しているため。

### Consequences

* Good, because symlink ベースの厳密な依存管理により、マイクロサービス間の依存分離が明確になる
* Good, because Vite + Fastify エコシステムとの親和性が高く、豊富なモノレポ構築事例を参考にできる
* Bad, because ビルドキャッシュやタスクオーケストレーションが必要になった場合、Turborepo 等の追加導入を検討する必要がある（ただし Vite+ の monorepo タスクランナーが代替候補となる可能性がある）

## More Information

* pnpm 公式ドキュメント: https://pnpm.io
* pnpm ワークスペース: https://pnpm.io/workspaces
