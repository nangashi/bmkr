---
status: "accepted"
date: 2026-03-24
last-validated: 2026-03-24
---

# EC サイト（React SPA）の CSS フレームワークに Tailwind CSS v4 を採用する

## Context and Problem Statement

EC サイト（React SPA）の CSS フレームワークが未選定であり、AI デザインツール（v0, Claude Artifacts）の出力をそのまま活用でき、デザイントークンを `/design` スキルから参照可能な技術を選定する必要がある。管理画面 UI（templ + htmx, ADR-0007）は本 ADR の対象外とする。

## Prerequisites

* React + Vite を採用済み (ADR-0001)
* pnpm モノレポを採用済み (ADR-0002)
* Oxlint + Oxfmt を採用済み — Tailwind クラスソート built-in (ADR-0010)
* #23 /design スキルがデザイントークンの管理場所を参照する
* AI デザインツール（v0, Claude Artifacts）の出力との互換性が必要

## Decision Drivers

* AI デザインツール出力との互換性 — v0, Claude Artifacts が生成するコードをそのまま利用できるか、変換コストはどの程度か
* デザイントークン管理の柔軟性 — /design スキルが参照するトークン（色・spacing 等）の定義・一元管理のしやすさ
* Vite + pnpm モノレポとのビルド統合 — HMR 速度、設定の共有・分離、ビルドパイプラインとの親和性
* エコシステム成熟度 — コンポーネントライブラリ（shadcn/ui 等）対応状況、ドキュメント、トラブルシュート情報の充実度

## Considered Options

* Tailwind CSS v4 + Vite プラグイン
* Tailwind CSS v3 + PostCSS

### Excluded Options

* UnoCSS — v0/Claude Artifacts が Tailwind コードしか生成せず、変換コストが大きいため除外
* PandaCSS — JSX スタイルプロップ方式で AI 出力と根本的に異なるため除外
* vanilla-extract — `.css.ts` 形式を AI ツールが生成しないため除外
* StyleX (Meta) — EC サイト SPA には過剰であり、AI ツール互換性がないため除外

## Comparison Overview

| 判断軸 | Tailwind v4 + Vite プラグイン | Tailwind v3 + PostCSS |
|--------|:-:|:-:|
| AI デザインツール出力との互換性 | ◎ v0 が v4 デフォルト化済み、ユーティリティクラスはほぼ共通 | △ v0 が v4 に移行、v3 出力の継続性リスク |
| デザイントークン管理の柔軟性 | ◎ `@theme` で CSS 変数自動生成、CSS `@import` で共有 | △ JS config のみ、CSS 変数は手動管理 |
| Vite + pnpm モノレポとのビルド統合 | ◎ 公式 Vite プラグイン、HMR <100ms、PostCSS 不要 | ○ PostCSS 経由、HMR ~500ms、安定動作 |
| エコシステム成熟度 | ○ shadcn/ui 完全対応、活発な開発、情報量は増加中 | ◎ 膨大な実績・情報量、shadcn/ui 引き続き対応、EOL 2027/02 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Tailwind CSS v4 + Vite プラグイン

CSS-first 設定（`@theme` ディレクティブ）と Rust 製 Oxide エンジンを特徴とする最新メジャーバージョン。公式 `@tailwindcss/vite` プラグインにより PostCSS 不要。

* Good, because v0 が v4 をデフォルト出力に切り替え済みで、AI 生成コードをそのまま利用できる
* Good, because `@theme` ディレクティブでデザイントークンが CSS カスタムプロパティとして自動公開され、`/design` スキルからの参照が容易
* Good, because CSS `@import` でテーマファイルを共有でき、モノレポでの設定管理が v3 の JS preset より簡潔
* Good, because 公式 `@tailwindcss/vite` プラグインで PostCSS 不要、Rust 製 Oxide エンジンにより HMR <100ms（v3 比 5 倍以上高速）
* Good, because shadcn/ui が v4 完全対応済み、OKLCH カラーに移行
* Good, because Oxfmt の Tailwind クラスソートが built-in で追加設定不要 (ADR-0010)
* Bad, because Claude Artifacts が v3 構文ベースのコードを生成する傾向があり、一部ユーティリティの名称変更（`shadow-sm` → `shadow-xs` 等）への対応が必要な場合がある
* Bad, because v4 はリリースから約 1 年で、v3 と比較するとトラブルシュート情報がまだ少ない

### Tailwind CSS v3 + PostCSS

`tailwind.config.ts` による JS ベースの設定と PostCSS プラグインを使用する従来方式。5 年以上の実績がある。

* Good, because 膨大な実績・チュートリアル・トラブルシュート情報がある
* Good, because Claude Artifacts の出力が v3 構文ベースのため、変換なしでそのまま利用できる可能性が高い
* Good, because shadcn/ui が v3 を引き続きサポートしており、既存コンポーネントがそのまま動作する
* Bad, because v0 が v4 デフォルトに移行済みのため、v0 出力を v3 プロジェクトで使うには変換が必要になる
* Bad, because デザイントークンが `tailwind.config.ts` 内の JS オブジェクトに閉じており、CSS カスタムプロパティとして自動公開されない
* Bad, because PostCSS 経由で HMR ~500ms、Oxide エンジン搭載の v4 と比較して開発体験が劣る
* Bad, because EOL が 2027/02/28 と見込まれ、新規プロジェクトでの採用は将来の v4 移行コストを確定させる

## Decision Outcome

Chosen option: "Tailwind CSS v4 + Vite プラグイン"

### Rationale

AI デザインツール出力との互換性を最重視した。v0 が v4 をデフォルト出力に切り替え済みであり、今後 AI ツールの v4 対応はさらに進む方向にある。加えて、`@theme` ディレクティブによるデザイントークンの CSS カスタムプロパティ自動公開は `/design` スキルとの連携に有利であり、公式 Vite プラグインによる HMR <100ms のビルド統合も v3 を上回る。エコシステム成熟度では v3 に劣るが、shadcn/ui の v4 完全対応により実用上の障壁は低いと判断した。

### Accepted Tradeoffs

* エコシステム成熟度の低さ — v3 と比較してトラブルシュート情報が少なく、問題発生時は GitHub Issues が主な情報源となる（比較表のエコシステム成熟度 ○ に対応）
* Claude Artifacts の v3 構文傾向 — AI 出力に v3 の旧ユーティリティ名が含まれる場合があり、手動での修正が必要な場面がある（比較表の AI 互換性 ◎ の留意点）

### Consequences

* Good, because v0 の出力をそのまま EC サイトの React コンポーネントに取り込める
* Good, because `@theme` で定義したデザイントークンが CSS カスタムプロパティとして自動公開され、`/design` スキルからの参照が容易になる
* Bad, because v4 固有の問題に遭遇した場合、情報が限られるため解決に時間がかかる可能性がある

## More Information

* Tailwind CSS v4 公式ドキュメント: https://tailwindcss.com/docs
* `@theme` ディレクティブ: https://tailwindcss.com/docs/theme
* shadcn/ui Tailwind v4 ガイド: https://ui.shadcn.com/docs/tailwind-v4
* Tailwind v3 からの移行ガイド: https://tailwindcss.com/docs/upgrade-guide
* Tailwind CSS v3 EOL 情報: https://endoflife.date/tailwind-css

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-24 | 初版作成 | N/A |
