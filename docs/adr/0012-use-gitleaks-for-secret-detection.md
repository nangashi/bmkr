---
status: "implemented"
date: 2026-03-18
last-validated: 2026-03-18
---

# シークレット検出ツールの選定

## Decision Outcome

検出精度とスキャン速度・導入容易さの2軸を最も重視し、Gitleaks を採用する。Gitleaks は Recall 88% で検出漏れが少なく、単一バイナリで依存なし・pre-commit hook が数行で設定可能という手軽さが学習用プロジェクトに適している。TruffleHog のライブ検証機能は魅力的だが、速度面のトレードオフがあり、pre-commit hook としての日常的な利用には Gitleaks の軽量さが勝る。Betterleaks は技術的に最も優れているが、リリース直後でエコシステムが未成熟なため、現時点での採用は見送った。

### Accepted Tradeoffs

- オリジナル作者（Zach Rice）の離脱によるメンテナンス体制の将来不安を受け入れる。Betterleaks の成熟を注視し、必要に応じて移行を検討する（`.gitleaks.toml` と CLI 互換のため移行コストは低い）
- Precision 46% の偽陽性に対し、allowlist の継続的なメンテナンスが必要になることを受け入れる

### Consequences

- pre-commit hook により、シークレットのコミットを開発段階で防止できる
- CI/CD パイプラインでの Git 履歴スキャンにより、過去の漏洩も検出可能
- Precision 46% の偽陽性に対し、allowlist の継続的なメンテナンスが必要になる

## Context and Problem Statement

本プロジェクト（Go + TypeScript/React のマイクロサービス構成）において、ソースコードや Git 履歴に含まれるシークレット（API キー、パスワード、トークン等）の漏洩を防止・検出するためのツールを選定する必要がある。pre-commit hook や CI/CD パイプラインでの利用を想定する。

## Prerequisites

- pnpm monorepo 構成 (ADR-0002)、Go と TypeScript の両方のコードベースが存在
- 学習用プロジェクトのため、OSS/無料で利用可能なツールが前提

## Decision Drivers

- 検出精度（偽陽性/偽陰性の少なさ）
- スキャン速度と導入の手軽さ
- カスタムルール・設定の柔軟性
- コミュニティの活発さとメンテナンス状況

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **Gitleaks（採用）** | Go 製の単一バイナリ。160+ ルールで主要サービスのシークレットを検出 |
| TruffleHog | 800+ 検出器とライブ検証機能を持つ。企業（Truffle Security Co.）がフルタイム開発 |
| detect-secrets | Yelp 主導の Python 製ツール。ベースライン方式で偽陽性を段階的に管理 |
| Trivy | Aqua Security 製の統合セキュリティスキャナ。シークレット検出は機能の一部 |
| Betterleaks | BPE トークン化 + CEL 検証式で Recall 98.6% を達成。Gitleaks 互換 |

除外: Titus — 新興プロジェクトで情報が少なすぎるため

## Comparison Overview

| 判断軸 | Gitleaks | TruffleHog | detect-secrets | Trivy | Betterleaks |
|--------|----------|------------|----------------|-------|-------------|
| 検出精度 | ◎ Recall 88%, 160+ルール | ◎ 800+検出器, ライブ検証 | ○ ベースライン方式で偽陽性抑制 | △ 専用ツールより検出力劣る | ◎ Recall 98.6%, BPEトークン化+CEL検証 |
| スキャン速度・導入容易さ | ◎ 単一バイナリ, 高速, pre-commit容易 | ○ 多彩なインストール, ライブ検証時は低速 | △ Python依存, セットアップに一手間 | ○ 単一バイナリだが重い | ○ Pure Go, 並列スキャン, ただし公式Action未整備 |
| カスタムルール・柔軟性 | ◎ .gitleaks.toml, Composite Rules | ○ config.yaml, カスタム検出器+検証EP | ○ カスタムプラグイン(Python) | △ シークレット特化設定は限定的 | ◎ .gitleaks.toml互換+CELによる高度な検証式 |
| コミュニティ・メンテナンス | ○ 作者離脱 | ◎ 企業主導で高頻度リリース | ○ Yelp主導で安定 | ◎ Aqua Security主導 | △ リリース直後, 作者の実績は確か |

## Notes

- TruffleHog のライセンスは AGPL-3.0（CLI 外部実行なら問題ないが、ライブラリ組込みには注意）
- TruffleHog はプロジェクトルートの規約設定ファイルがなく `--config` で明示指定が必要
- detect-secrets には Git 履歴のスキャン機能がネイティブにはない（ファイルシステムスキャン中心）
- Betterleaks は 2026-03-12 リリースで公式 GitHub Action がまだなく、CI 統合は CLI 直接実行が必要

## More Information

* [Gitleaks GitHub](https://github.com/gitleaks/gitleaks)
* [Betterleaks GitHub](https://github.com/betterleaks/betterleaks) — 将来の移行先候補
* [A Comparative Study of Software Secrets Reporting by Secret Detection Tools](https://arxiv.org/pdf/2307.00714) — Recall/Precision の根拠
