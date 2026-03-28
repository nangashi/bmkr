---
status: "implemented"
date: 2026-03-17
last-validated: 2026-03-17
---

# Go の formatter/linter の選定

## Decision Outcome

Go エコシステムでの標準度を最も重視し、golangci-lint v2 を採用する。Kubernetes・Terraform・Connect-RPC 等の主要 OSS が採用するデファクトスタンダードであり、学んだ設定・運用スキルがそのまま実務に直結する点が、学習目的のプロジェクトにおいて最も価値が高いと判断した。`linters.default: standard` で合理的なデフォルトセットから始められるため、設定の複雑さを段階的にコントロールできる。lint ルールの網羅性でも最も優れており、formatter 統合（`golangci-lint fmt`）により単一ツールで完結する運用のシンプルさも決め手となった。

### Accepted Tradeoffs

- 初回実行の遅さ（キャッシュなし約50秒）およびメモリ消費の増大を受け入れる。学習用プロジェクトの規模では実用上問題にならないと判断した（比較表の「実行速度」で ○ 評価）
- 100+ linter の取捨選択・設定チューニングに学習コストがかかることを受け入れる。ただし `linters.default: standard` による段階的導入で緩和する

### Consequences

- 実務でも通用するデファクトスタンダードのツールの運用スキルが身につく
- formatter（gofumpt + gci）も `golangci-lint fmt` で一元管理でき、ツール管理が単一で済む
- sqlc/protobuf 生成コードが `generated: strict` で自動除外され、運用負荷が低い
- 100+ linter の取捨選択・設定チューニングに学習コストがかかる

## Context and Problem Statement

学習用マイクロサービスプロジェクトの Go サービス（ec-site、product-mgmt、customer-mgmt）に formatter/linter を導入し、コード品質の維持と Go のベストプラクティスの学習を促進したい。protobuf 生成コード（gen/go/）や sqlc 生成コード（db/generated/）は自動生成のため lint 対象外とする。

## Prerequisites

- バックエンド API は Go 1.26.1 を使用、Echo フレームワーク採用済み (ADR-0001)
- Connect-RPC でサービス間通信 (ADR-0003)、sqlc でデータベースアクセスコード自動生成 (ADR-0004)
- 学習目的のプロジェクトである

## Decision Drivers

- lint ルールの網羅性
- Go エコシステムでの標準度
- 実行速度
- 設定・運用のシンプルさ

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **golangci-lint v2（採用）** | 100+ linter + 6 formatter を単一ツール・単一設定ファイルで一元管理する統合型 |
| gofmt + staticcheck | Go 標準 formatter + 深い静的解析を組み合わせた標準ツール型 |
| gofmt + revive | Go 標準 formatter + golint 後継の軽量カスタマイズ型 |

## Comparison Overview

| 判断軸 | golangci-lint v2 | gofmt + staticcheck | gofmt + revive |
|--------|-----------------|---------------------|----------------|
| lint ルールの網羅性 | ◎ 100+ linter 統合。バグ・スタイル・セキュリティ・複雑度を全方位カバー | ○ 150+ チェックで深い静的解析。ただしセキュリティ・複雑度は対象外 | △ 127 ルールだがセキュリティ・深い静的解析が弱い。go vet 別途必要 |
| Go エコシステムでの標準度 | ◎ K8s・Terraform 等が採用。CI テンプレートのデファクト | ○ VS Code デフォルト・gopls 統合。ただし CI は golangci-lint 経由が主流 | △ golint 後継だが単体採用は少数派 |
| 実行速度 | ○ 並列実行+キャッシュで実用的。キャッシュあり約14秒。`--fast-only` で高速化可 | ○ gofmt は極めて高速。staticcheck もキャッシュ活用で2回目以降は高速 | ◎ golint 比6倍高速。型チェック不要ルールのみなら特に高速。軽量 |
| 設定・運用のシンプルさ | ○ 単一 `.golangci.yml` で全管理。ただし100+ linter の設定は肥大化しうる | △ 2ツール管理。staticcheck の生成コード除外に手間。将来の golangci-lint 移行時に設定二重管理リスク | ○ TOML 1ファイル + goimports。ただし go vet 別途必要で計3ツール管理 |

## Notes

- golangci-lint のライセンスは GPL-3.0（CI 実行のみなら通常問題なし）
- staticcheck は単一メンテナー（Dominik Honnef）でバス因子が低い
- staticcheck には自動生成コードの除外機能が弱い問題がある（issue #429 未解決）

## More Information

* golangci-lint 公式ドキュメント: https://golangci-lint.run/
* golangci-lint v2 移行ガイド: https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/
* golangci-lint GitHub Action: https://github.com/golangci/golangci-lint-action
* golangci-lint GitHub リポジトリ: https://github.com/golangci/golangci-lint
