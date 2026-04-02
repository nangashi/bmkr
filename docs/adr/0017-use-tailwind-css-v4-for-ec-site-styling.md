---
status: "accepted"
date: 2026-03-24
last-validated: 2026-03-24
---

# EC サイト（React SPA）の CSS フレームワークの選定

## Decision Outcome

Tailwind CSS v4 + Vite プラグインを採用する。v0 が v4 をデフォルト出力に切り替え済みであり、今後 AI ツールの v4 対応はさらに進む方向にある。`@theme` ディレクティブによるデザイントークンの CSS カスタムプロパティ自動公開は `/design` スキルとの連携に有利であり、公式 Vite プラグインによる HMR の高速なビルド統合も v3 を上回る。エコシステム成熟度では v3 に劣るが、shadcn/ui の v4 完全対応により実用上の障壁は低いと判断した。

### Accepted Tradeoffs

- エコシステム成熟度の低さ — v3 と比較してトラブルシュート情報が少なく、問題発生時は GitHub Issues が主な情報源となる
- Claude Artifacts の v3 構文傾向 — AI 出力に v3 の旧ユーティリティ名が含まれる場合があり、手動での修正が必要な場面がある

### Consequences

- v0 の出力をそのまま EC サイトの React コンポーネントに取り込める
- `@theme` で定義したデザイントークンが CSS カスタムプロパティとして自動公開され、`/design` スキルからの参照が容易になる
- v4 固有の問題に遭遇した場合、情報が限られるため解決に時間がかかる可能性がある

## Context and Problem Statement

EC サイト（React SPA）の CSS フレームワークが未選定であり、AI デザインツール（v0, Claude Artifacts）の出力をそのまま活用でき、デザイントークンを `/design` スキルから参照可能な技術を選定する必要がある。管理画面 UI（templ + htmx, ADR-0007）は本 ADR の対象外とする。

## Prerequisites

- React + Vite、pnpm モノレポを採用済み (ADR-0001, ADR-0002)
- Oxlint + Oxfmt を採用済み — Tailwind クラスソート built-in (ADR-0010)
- #23 /design スキルがデザイントークンの管理場所を参照する
- AI デザインツール（v0, Claude Artifacts）の出力との互換性が必要

## Decision Drivers

- AI デザインツール出力との互換性
- デザイントークン管理の柔軟性
- Vite + pnpm モノレポとのビルド統合
- エコシステム成熟度

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **Tailwind CSS v4 + Vite プラグイン（採用）** | CSS-first 設定と Rust 製 Oxide エンジンによる最新メジャーバージョン。PostCSS 不要 |
| Tailwind CSS v3 + PostCSS | JS ベースの設定と PostCSS プラグインを使用する従来方式 |

除外: UnoCSS（AI ツールが Tailwind コードしか生成せず変換コスト大）、PandaCSS（JSX スタイルプロップ方式で AI 出力と不一致）、vanilla-extract（AI ツールが `.css.ts` を生成しない）、StyleX（EC サイト SPA には過剰で AI ツール互換性なし）

## Comparison Overview

| 判断軸 | Tailwind v4 + Vite プラグイン | Tailwind v3 + PostCSS |
|--------|:-:|:-:|
| AI デザインツール出力との互換性 | ◎ v0 が v4 デフォルト化済み、ユーティリティクラスはほぼ共通 | △ v0 が v4 に移行、v3 出力の継続性リスク |
| デザイントークン管理の柔軟性 | ◎ `@theme` で CSS 変数自動生成、CSS `@import` で共有 | △ JS config のみ、CSS 変数は手動管理 |
| Vite + pnpm モノレポとのビルド統合 | ◎ 公式 Vite プラグイン、高速 HMR、PostCSS 不要 | ○ PostCSS 経由、HMR は v4 比で低速、安定動作 |
| エコシステム成熟度 | ○ shadcn/ui 完全対応、活発な開発、情報量は増加中 | ◎ 膨大な実績・情報量、shadcn/ui 引き続き対応、EOL 2027/02 |

## Notes

- Tailwind v3 の EOL は 2027/02/28 と見込まれ、新規プロジェクトでの採用は将来の v4 移行コストを確定させる
- v3 はデザイントークンが `tailwind.config.ts` 内の JS オブジェクトに閉じており、CSS カスタムプロパティとして自動公開されない

## More Information

* Tailwind CSS v4 公式ドキュメント: https://tailwindcss.com/docs
* `@theme` ディレクティブ: https://tailwindcss.com/docs/theme
* shadcn/ui Tailwind v4 ガイド: https://ui.shadcn.com/docs/tailwind-v4
* Tailwind v3 からの移行ガイド: https://tailwindcss.com/docs/upgrade-guide
* Tailwind CSS v3 EOL 情報: https://endoflife.date/tailwind-css
