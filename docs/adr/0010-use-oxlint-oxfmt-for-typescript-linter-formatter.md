---
status: "implemented"
date: 2026-03-17
last-validated: 2026-03-17
---

# TypeScript の linter/formatter の選定

## Decision Outcome

実行速度を最重視し、Oxlint + Oxfmt (Beta) を採用する。全工程 Rust ベースで lint 50-100x / format 30-35x（Prettier 比）の最速構成であり、CI・エディタ双方での開発体験が最も良い。VoidZero 社が Vite+ の中核ツールに位置づけており、今後の業界採用拡大を見込んで先行投資する判断とした。同一 VSCode 拡張で lint + format が完結する点も、統合成熟度の観点で評価した。

### Accepted Tradeoffs

- Oxfmt が Beta であるため、アップデート時に破壊的変更への対応が必要になる可能性がある（比較表の実務汎用性 △ に対応）
- type-aware linting が Alpha 段階であり、大規模コードベースではメモリ問題や精度の課題が残る（比較表のルール充実度 ○ に対応）
- コミュニティ情報が ESLint + Prettier と比較して少なく、トラブルシュート時は GitHub Issues が主な情報源となる

### Consequences

- Vite+ エコシステムの中核ツールを先取りして学べる
- 同一 VSCode 拡張で lint + format が完結し、開発体験が統一される
- トラブル時の情報源が GitHub Issues に限られ、解決に時間がかかる可能性がある

## Context and Problem Statement

BFF (Fastify) とフロントエンド (React/Vite) の TypeScript コードに対して linter/formatter が未設定であり、コード品質と一貫性を担保する仕組みを導入する必要がある。本 ADR は TypeScript プロジェクトの lint/format ツール選定を対象とし、Go サービスの lint/format は対象外とする。

## Prerequisites

- React (Vite) + Fastify (TypeScript) を採用済み (ADR-0001)、pnpm モノレポ構成 (ADR-0002)
- 学習目的のプロジェクトである

## Decision Drivers

- Linter としてのルール充実度
- 実行速度
- Fastify / React (Vite) との統合成熟度
- 実務での汎用性

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| ESLint + Prettier | 業界で最も広く採用されている lint + format の組み合わせ |
| Biome | Rust 製の lint + format 統合ツール。1つの設定ファイルで完結 |
| Oxlint + Prettier | VoidZero 社による Rust 製 linter + 業界標準 formatter の組み合わせ |
| **Oxlint + Oxfmt (Beta)（採用）** | VoidZero 社/Oxc プロジェクトによる Rust 製 linter + formatter。全工程 Rust ベースで最速 |

除外: Oxlint + Biome (formatter only) — 同一エコシステム（VoidZero/Oxc）に Oxfmt が存在し、採用事例が確認できなかったため

## Comparison Overview

| 判断軸 | ESLint + Prettier | Biome | Oxlint + Prettier | Oxlint + Oxfmt |
|--------|:-:|:-:|:-:|:-:|
| Linter ルール充実度 | ◎ コア200+/プラグイン1000+/type-aware完全 | ○ 459ルール/type-aware 75-85% | ○ 695+/type-aware alpha (59/61) | ○ 695+/type-aware alpha (59/61) |
| 実行速度 | △ 基準 | ○ ESLint比10-25x | ○ lint 50-100x / format基準 | ◎ lint 50-100x / format 30-35x |
| Fastify/React(Vite)統合 | ◎ 最も成熟 | ○ VSCode拡張にやや不安定報告 | ○ VoidZero製+成熟Prettier | ◎ 同一VSCode拡張で完結/Vite+中核 |
| 実務での汎用性 | ◎ デファクト標準 | ○ 成長中 | ○ Oxlint採用実績あり+Prettier | △ Oxfmt Beta/採用は先進企業のみ |

## Notes

- Oxfmt は Prettier conformance 100%（JS/TS）で、`oxfmt --migrate prettier` によるワンコマンド移行が可能
- Oxfmt のデフォルト設定は Prettier と異なる（printWidth: 100, trailingComma: "all"）
- Oxfmt はネスト設定ファイル未対応（モノレポでパッケージ別設定ができない）
- import ソート・Tailwind CSS クラスソート・package.json ソートが Oxfmt に built-in
- Biome の VSCode 拡張に安定性の問題報告あり（LSP クラッシュ、pnpm モノレポでのバイナリ検出失敗）

## More Information

* Oxlint 公式ドキュメント: https://oxc.rs/docs/guide/usage/linter.html
* Oxfmt 公式ドキュメント: https://oxc.rs/docs/guide/usage/formatter.html
* Prettier からの移行ガイド: https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier
* VSCode 拡張: https://marketplace.visualstudio.com/items?itemName=oxc.oxc-vscode
* Vite+ 構想: https://voidzero.dev/posts/announcing-vite-plus
