---
status: "accepted"
date: 2026-03-17
---

# TypeScript の linter/formatter に Oxlint + Oxfmt を採用する

## Context and Problem Statement

BFF (Fastify) とフロントエンド (React/Vite) の TypeScript コードに対して linter/formatter が未設定であり、コード品質と一貫性を担保する仕組みを導入する必要がある。本 ADR は TypeScript プロジェクトの lint/format ツール選定を対象とし、Go サービスの lint/format は対象外とする。

## Prerequisites

* React (Vite) + Fastify (TypeScript) を採用済み (ADR-0001)
* pnpm モノレポを採用済み (ADR-0002)
* 学習目的のプロジェクトである

## Decision Drivers

* Linter としてのルール充実度 — 対応ルール数、型情報を活用した lint（type-aware linting）の対応状況、React/Fastify 向けプラグインの有無
* 実行速度 — lint/format の処理速度、CI やエディタでの体感速度
* Fastify / React (Vite) との統合成熟度 — 公式プラグイン・推奨設定の有無、IDE（VSCode 等）との連携の安定性
* 実務での汎用性 — 業務プロジェクトでの採用実績、求人・コミュニティでの認知度

## Considered Options

* ESLint + Prettier
* Biome
* Oxlint + Prettier
* Oxlint + Oxfmt (Beta)

### Excluded Options

* Oxlint + Biome (formatter only) — 同一エコシステム（VoidZero/Oxc）に Oxfmt が存在し、採用事例が確認できなかった（0件）ため除外

## Comparison Overview

| 判断軸 | ESLint + Prettier | Biome | Oxlint + Prettier | Oxlint + Oxfmt |
|--------|:-:|:-:|:-:|:-:|
| Linter ルール充実度 | ◎ コア200+/プラグイン1000+/type-aware完全 | ○ 459ルール/type-aware 75-85% | ○ 695+/type-aware alpha (59/61) | ○ 695+/type-aware alpha (59/61) |
| 実行速度 | △ 基準 | ○ ESLint比10-25x | ○ lint 50-100x / format基準 | ◎ lint 50-100x / format 30-35x |
| Fastify/React(Vite)統合 | ◎ 最も成熟 | ○ VSCode拡張にやや不安定報告 | ○ VoidZero製+成熟Prettier | ◎ 同一VSCode拡張で完結/Vite+中核 |
| 実務での汎用性 | ◎ デファクト標準 | ○ 成長中(550万DL/週) | ○ Oxlint採用実績あり+Prettier | △ Oxfmt Beta/採用は先進企業のみ |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### ESLint + Prettier

業界で最も広く採用されている lint + format の組み合わせ（ESLint 27.2K Stars, Prettier 51.8K Stars）。

* Good, because ルール数が圧倒的（コア200+ / プラグイン1000+）、typescript-eslint の type-aware ルール約60個が完全動作
* Good, because Vite の react-ts テンプレートに ESLint 同梱済み、eslint-plugin-react-hooks は React 公式メンテナンス
* Good, because npm 週間 8,500万DL、求人市場のデファクトスタンダード
* Good, because ESLint v10 (2026/02) でモノレポ設定探索が安定化、マルチスレッド lint で 30-300% 高速化
* Bad, because Biome 比 10-25倍、Oxlint 比 50-100倍遅い（type-aware 有効時はさらに低速）
* Bad, because 2ツール構成で設定ファイル複数（eslint.config.js + .prettierrc）、Prettier 競合回避設定が必要

### Biome

Rust 製の lint + format 統合ツール（23.8K Stars）。1つの設定ファイルで完結。

* Good, because 459ルール内蔵、22の ESLint プラグインソースからルール移植済み（ESLint ルールの約80%カバー）
* Good, because lint + format が `biome.json` 1ファイルで完結、依存も1パッケージのみ
* Good, because ESLint の 10-25倍高速（10kファイル lint: 0.8秒 / format: 0.3秒）
* Good, because v2 で type-aware linting (Biotype) 導入、TypeScript コンパイラ非依存の独自型推論
* Good, because Vercel が社内標準として採用・スポンサー
* Bad, because type-aware linting の精度は typescript-eslint 比で約75-85%
* Bad, because VSCode 拡張に安定性の問題報告あり（LSP クラッシュ、pnpm モノレポでのバイナリ検出失敗）
* Bad, because ESLint のプラグインエコシステム（1000+）には及ばない

### Oxlint + Prettier

VoidZero 社による Rust 製 linter（16.5K Stars）+ 業界標準 formatter の組み合わせ。

* Good, because 695+ ルール内蔵（Biome の 459 を上回る）、React/react-hooks/jsx-a11y 等がビルトイン
* Good, because ESLint 比 50-100倍高速（メモリ 92MB vs 1.4GB）
* Good, because VoidZero 社が開発、Vite+ 統合ツールチェインの中核
* Good, because JS Plugins (Alpha) で既存 ESLint プラグインをそのまま実行可能
* Good, because Shopify, Airbnb, Mercedes-Benz 等で採用実績あり
* Good, because Prettier は成熟しており formatter 側の安定性リスクがない
* Bad, because type-aware linting は Alpha（tsgo/TS 7.0+ ベース）
* Bad, because Prettier との 2ツール/2拡張構成

### Oxlint + Oxfmt (Beta)

VoidZero 社/Oxc プロジェクトによる Rust 製 linter + formatter の組み合わせ。全工程 Rust ベースで最速。

* Good, because lint 50-100x + format 30-35x（Prettier比）で全工程最速。Biome formatter 比でも約3倍高速
* Good, because 同一 VSCode 拡張（oxc.oxc-vscode）で lint + format が完結、設定競合なし
* Good, because Vite+ 構想の中核（`vite lint` / `vite fmt` として統合予定）
* Good, because Prettier conformance 100%（JS/TS）、`oxfmt --migrate prettier` でワンコマンド移行可能
* Good, because import ソート・Tailwind CSS クラスソート・package.json ソートが built-in
* Good, because vuejs/core, vercel/turborepo, getsentry/sentry-javascript が採用済み
* Bad, because Oxfmt は Beta（破壊的変更の可能性）、type-aware linting は Alpha
* Bad, because デフォルト設定が Prettier と異なる（printWidth: 100, trailingComma: "all"）
* Bad, because ネスト設定ファイル未対応（モノレポでパッケージ別設定ができない）
* Bad, because コミュニティ情報が少なく、トラブル時は GitHub Issues が主な情報源

## Decision Outcome

Chosen option: "Oxlint + Oxfmt (Beta)"

### Rationale

実行速度を最重視した。全工程 Rust ベースで lint 50-100x / format 30-35x（Prettier 比）の最速構成であり、CI・エディタ双方での開発体験が最も良い。VoidZero 社が Vite+ の中核ツールに位置づけており、今後の業界採用拡大を見込んで先行投資する判断とした。同一 VSCode 拡張で lint + format が完結する点も、統合成熟度の観点で評価した。

### Accepted Tradeoffs

* Oxfmt が Beta であるため、アップデート時に破壊的変更への対応が必要になる可能性がある（比較表の実務汎用性 △ に対応）
* type-aware linting が Alpha 段階であり、大規模コードベースではメモリ問題や精度の課題が残る（比較表のルール充実度 ○ に対応）
* コミュニティ情報が ESLint + Prettier と比較して少なく、トラブルシュート時は GitHub Issues が主な情報源となる

### Consequences

* Good, because Vite+ エコシステムの中核ツールを先取りして学べる
* Good, because 同一 VSCode 拡張で lint + format が完結し、開発体験が統一される
* Bad, because Beta/Alpha 機能に依存するため、アップデート時に破壊的変更への対応が必要になる可能性がある
* Bad, because トラブル時の情報源が GitHub Issues に限られ、解決に時間がかかる可能性がある

## More Information

* Oxlint 公式ドキュメント: https://oxc.rs/docs/guide/usage/linter.html
* Oxfmt 公式ドキュメント: https://oxc.rs/docs/guide/usage/formatter.html
* Prettier からの移行ガイド: https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier
* VSCode 拡張: https://marketplace.visualstudio.com/items?itemName=oxc.oxc-vscode
* Vite+ 構想: https://voidzero.dev/posts/announcing-vite-plus
