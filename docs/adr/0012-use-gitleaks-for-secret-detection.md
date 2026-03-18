---
status: "accepted"
date: 2026-03-18
---

# シークレット検出ツールに Gitleaks を使用する

## Context and Problem Statement

本プロジェクト（Go + TypeScript/React のマイクロサービス構成）において、ソースコードや Git 履歴に含まれるシークレット（API キー、パスワード、トークン等）の漏洩を防止・検出するためのツールを選定する必要がある。pre-commit hook や CI/CD パイプラインでの利用を想定する。

## Prerequisites

* pnpm monorepo 構成 (ADR-0002)
* Go と TypeScript の両方のコードベースが存在
* 学習用プロジェクトのため、OSS/無料で利用可能なツールが前提

## Decision Drivers

* 検出精度（偽陽性/偽陰性の少なさ）
* スキャン速度と導入の手軽さ
* カスタムルール・設定の柔軟性
* コミュニティの活発さとメンテナンス状況

## Considered Options

* Gitleaks
* TruffleHog
* detect-secrets
* Trivy
* Betterleaks

### Excluded Options

* Titus — 新興プロジェクト（GitHub Stars 411）で情報が少なすぎるため評価対象から除外

## Comparison Overview

| 判断軸 | Gitleaks | TruffleHog | detect-secrets | Trivy | Betterleaks |
|--------|----------|------------|----------------|-------|-------------|
| 検出精度 | ◎ Recall 88%, 160+ルール | ◎ 800+検出器, ライブ検証 | ○ ベースライン方式で偽陽性抑制 | △ 専用ツールより検出力劣る | ◎ Recall 98.6%, BPEトークン化+CEL検証 |
| スキャン速度・導入容易さ | ◎ 単一バイナリ, 高速, pre-commit容易 | ○ 多彩なインストール, ライブ検証時は低速 | △ Python依存, セットアップに一手間 | ○ 単一バイナリだが重い | ○ Pure Go, 並列スキャン, ただし公式Action未整備 |
| カスタムルール・柔軟性 | ◎ .gitleaks.toml, Composite Rules | ○ config.yaml, カスタム検出器+検証EP | ○ カスタムプラグイン(Python) | △ シークレット特化設定は限定的 | ◎ .gitleaks.toml互換+CELによる高度な検証式 |
| コミュニティ・メンテナンス | ○ 25.4k Stars, 作者離脱 | ◎ 25.1k Stars, 企業主導で3-5日毎リリース | ○ 4.4k Stars, Yelp主導で安定 | ◎ 32.2k Stars, Aqua Security主導 | △ 436 Stars, リリース直後, 作者の実績は確か |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### Gitleaks

* Good, because Recall 88%と高い検出率で、160+のデフォルトルールが主要サービスを網羅
* Good, because 単一バイナリで依存なし、pre-commit hook 設定が数行で完了
* Good, because `.gitleaks.toml` による柔軟な設定（allowlist, Composite Rules, パス除外）
* Good, because MIT ライセンスで制約なし
* Bad, because Precision 46%で偽陽性が多め（ライブ検証機能なし）
* Bad, because オリジナル作者が離脱し、今後のメンテナンス体制に不確実性

### TruffleHog

* Good, because 800+検出器とライブ検証により、検出シークレットの有効/無効を自動判定できる
* Good, because 企業（Truffle Security Co.）がフルタイム開発、3-5日毎のリリースで活発
* Good, because `--results=verified,unknown` で偽陽性を大幅に削減可能
* Bad, because ライブ検証を含むフルスキャンは Gitleaks より低速
* Bad, because AGPL-3.0 ライセンス（CLI 外部実行なら問題ないが、ライブラリ組込みには注意）
* Bad, because プロジェクトルートの規約設定ファイルがなく `--config` で明示指定が必要

### detect-secrets

* Good, because ベースライン方式により既存コードの偽陽性を段階的に管理できる
* Good, because Python プラグインによるカスタム検出器の拡張が柔軟
* Bad, because Python ランタイムが必要で、Go+TS プロジェクトに追加の依存を持ち込む
* Bad, because スキャン速度が Go 製ツールより劣る
* Bad, because Git 履歴のスキャン機能がネイティブにはない（ファイルシステムスキャン中心）

### Trivy

* Good, because コンテナ脆弱性・IaC・ライセンス・シークレットを1ツールでカバー
* Good, because 32.2k Stars, Aqua Security による強力なメンテナンス
* Bad, because シークレット検出は「おまけ」的な位置づけで、専用ツールより検出精度が劣る
* Bad, because pre-commit hook としての利用は想定されておらず、CI/ファイルシステムスキャン向け
* Bad, because シークレット検出のカスタマイズ性は専用ツールに比べて限定的

### Betterleaks

* Good, because BPE トークン化により Recall 98.6% と最高の検出率
* Good, because CEL（Common Expression Language）による検証式で、正規表現より表現力の高いルール定義が可能
* Good, because Gitleaks の `.gitleaks.toml` と CLI オプションがそのまま動作し、移行コストが極めて低い
* Good, because Pure Go（CGO 不要）で並列スキャンにより高速化
* Bad, because 2026-03-12 リリースでコミュニティ規模が小さく（Stars 436）、実運用実績が乏しい
* Bad, because 公式 GitHub Action がまだなく、CI 統合は CLI 直接実行が必要

## Decision Outcome

Chosen option: "Gitleaks", because 検出精度（Recall 88%）とスキャン速度・導入容易さを重視した結果、最もバランスが良い。

### Rationale

検出精度とスキャン速度・導入容易さの2軸を最も重視した。Gitleaks は Recall 88% で検出漏れが少なく、単一バイナリで依存なし・pre-commit hook が数行で設定可能という手軽さが学習用プロジェクトに適している。TruffleHog のライブ検証機能は魅力的だが、速度面のトレードオフがあり、pre-commit hook としての日常的な利用には Gitleaks の軽量さが勝る。Betterleaks は技術的に最も優れているが、リリース直後でエコシステムが未成熟なため、現時点での採用は見送った。

### Accepted Tradeoffs

* オリジナル作者（Zach Rice）の離脱によるメンテナンス体制の将来不安を受け入れる。Betterleaks の成熟を注視し、必要に応じて移行を検討する（`.gitleaks.toml` と CLI 互換のため移行コストは低い）
* Precision 46% の偽陽性に対し、allowlist の継続的なメンテナンスが必要になることを受け入れる

### Consequences

* Good, because pre-commit hook により、シークレットのコミットを開発段階で防止できる
* Good, because CI/CD パイプラインでの Git 履歴スキャンにより、過去の漏洩も検出可能
* Bad, because Precision 46% の偽陽性に対し、allowlist の継続的なメンテナンスが必要になる

## More Information

* [Gitleaks GitHub](https://github.com/gitleaks/gitleaks)
* [Betterleaks GitHub](https://github.com/betterleaks/betterleaks) — 将来の移行先候補
* [A Comparative Study of Software Secrets Reporting by Secret Detection Tools](https://arxiv.org/pdf/2307.00714) — Recall/Precision の根拠
