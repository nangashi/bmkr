---
status: "accepted"
date: 2026-03-17
---

# Go の formatter/linter に golangci-lint v2 を採用する

## Context and Problem Statement

学習用マイクロサービスプロジェクトの Go サービス（ec-site、product-mgmt、customer-mgmt）に formatter/linter を導入し、コード品質の維持と Go のベストプラクティスの学習を促進したい。protobuf 生成コード（gen/go/）や sqlc 生成コード（db/generated/）は自動生成のため lint 対象外とする。

## Prerequisites

* バックエンド API は Go 1.26.1 を使用 (ADR-0001)
* Echo フレームワークを採用済み (ADR-0001)
* Connect-RPC でサービス間通信 (ADR-0003)
* sqlc でデータベースアクセスコード自動生成 (ADR-0004)
* 学習目的のプロジェクトである

## Decision Drivers

* lint ルールの網羅性 — バグ検出・スタイル統一・パフォーマンス改善など、どれだけ幅広い観点でコードをチェックできるか
* Go エコシステムでの標準度 — 実務プロジェクトや OSS での採用率、CI テンプレートでのデフォルト採用状況
* 実行速度 — lint/format の実行にかかる時間、キャッシュ機構の有無、大規模プロジェクトでのスケーラビリティ
* 設定・運用のシンプルさ — 初期セットアップの容易さ、設定ファイルの複雑さ、マルチサービスでの設定共有のしやすさ

## Considered Options

* golangci-lint v2（統合型）
* gofmt + staticcheck（標準ツール組み合わせ型）
* gofmt + revive（軽量カスタマイズ型）

### Excluded Options

なし

## Comparison Overview

| 判断軸 | golangci-lint v2 | gofmt + staticcheck | gofmt + revive |
|--------|-----------------|---------------------|----------------|
| lint ルールの網羅性 | ◎ 100+ linter 統合。バグ・スタイル・セキュリティ・複雑度を全方位カバー | ○ 150+ チェックで深い静的解析。ただしセキュリティ・複雑度は対象外 | △ 127 ルールだがセキュリティ・深い静的解析が弱い。go vet 別途必要 |
| Go エコシステムでの標準度 | ◎ Stars 18,679。K8s・Terraform 等が採用。CI テンプレートのデファクト | ○ Stars 6,700。VS Code デフォルト・gopls 統合。ただし CI は golangci-lint 経由が主流 | △ Stars 5,400。golint 後継だが単体採用は少数派 |
| 実行速度 | ○ 並列実行+キャッシュで実用的。キャッシュあり約14秒。`--fast-only` で高速化可 | ○ gofmt は極めて高速。staticcheck もキャッシュ活用で2回目以降は高速 | ◎ golint 比6倍高速。型チェック不要ルールのみなら特に高速。軽量 |
| 設定・運用のシンプルさ | ○ 単一 `.golangci.yml` で全管理。ただし100+ linter の設定は肥大化しうる | △ 2ツール管理。staticcheck の生成コード除外に手間。将来の golangci-lint 移行時に設定二重管理リスク | ○ TOML 1ファイル + goimports。ただし go vet 別途必要で計3ツール管理 |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### golangci-lint v2（統合型）

* Good, because 100+ linter + 6 formatter を単一ツール・単一設定ファイルで一元管理でき、lint ルールの網羅性が最も高い
* Good, because Go エコシステムのデファクトスタンダード（Stars 18,679）で、K8s・Terraform・Connect-RPC 等の主要 OSS が採用
* Good, because v2 の `golangci-lint fmt` で formatter も統合され、gofumpt + gci を別途管理する必要がない
* Good, because sqlc/protobuf の生成コードを `generated: strict` で自動除外でき、運用が楽
* Good, because `--fast-only` や `linters.default: standard` で段階的導入が容易
* Good, because GitHub Action 公式対応（v9）、主要 IDE 統合済み
* Bad, because linter を大量に有効化すると初回実行が遅く（キャッシュなし約50秒）、メモリ消費も増大する
* Bad, because 100+ linter のフルカスタマイズは設定ファイルが肥大化し、学習コストがかかる
* Bad, because ライセンスが GPL-3.0（CI 実行のみなら通常問題なし）

### gofmt + staticcheck（標準ツール組み合わせ型）

* Good, because SA カテゴリ（90+ チェック）の深い静的解析は他の linter にない精度で、実際のバグを検出できる
* Good, because false positive が極めて少ない設計で、警告の信頼性が高い
* Good, because VS Code のデフォルト linter・gopls 統合済みで、エディタ上でリアルタイムフィードバック
* Good, because `-explain` フラグで各チェックの詳細を学べ、教育的価値が高い
* Good, because ライセンスが MIT + BSD-3-Clause で制約なし
* Bad, because セキュリティ（gosec 相当）・複雑度（gocyclo 相当）・エラーハンドリング（errcheck 相当）のチェックがなく、単体では網羅性に欠ける
* Bad, because 自動生成コードの除外機能が弱い（issue #429 未解決）。sqlc/protobuf 生成コードへの対処に手間がかかる
* Bad, because 単一メンテナー（Dominik Honnef）でバス因子が低い

### gofmt + revive（軽量カスタマイズ型）

* Good, because 127 ルール中、複雑度メトリクス（cyclomatic, cognitive-complexity, function-length）が充実しており、コード品質の定量評価に強い
* Good, because golint 比6倍高速（型チェック無効時）。3選択肢中で最も軽量・高速
* Good, because TOML 設定がシンプルで、ルール単位のファイル除外パターン（glob）が柔軟
* Good, because MIT ライセンスで制約なし。メンテナンスも非常に活発
* Good, because カスタムルールを Go で自作できるフレームワークを持つ
* Bad, because Go エコシステムでの単体採用は少数派。CI テンプレートのデフォルトは golangci-lint
* Bad, because `go vet` 相当のチェックが含まれず、別途実行が必要（計3ツール管理）
* Bad, because セキュリティチェック・深い静的解析（nil dereference 等）は含まれない

## Decision Outcome

Chosen option: "golangci-lint v2（統合型）", because Go エコシステムでの標準度を重視し、`linters.default: standard` で導入する。

### Rationale

Go エコシステムでの標準度を最も重視した。golangci-lint v2 は Stars 18,679 を誇り、Kubernetes・Terraform・Connect-RPC 等の主要 OSS が採用するデファクトスタンダードである。学んだ設定・運用スキルがそのまま実務に直結する点が、学習目的のプロジェクトにおいて最も価値が高いと判断した。また `linters.default: standard` で合理的なデフォルトセットから始められるため、設定の複雑さを段階的にコントロールできる。lint ルールの網羅性でも最も優れており、formatter 統合（`golangci-lint fmt`）により単一ツールで完結する運用のシンプルさも決め手となった。

### Accepted Tradeoffs

* 初回実行の遅さ（キャッシュなし約50秒）およびメモリ消費の増大を受け入れる。学習用プロジェクトの規模では実用上問題にならないと判断した（比較表の「実行速度」で ○ 評価）
* 100+ linter の取捨選択・設定チューニングに学習コストがかかることを受け入れる。ただし `linters.default: standard` による段階的導入で緩和する

### Consequences

* Good, because 実務でも通用するデファクトスタンダードのツールの運用スキルが身につく
* Good, because formatter（gofumpt + gci）も `golangci-lint fmt` で一元管理でき、ツール管理が単一で済む
* Good, because sqlc/protobuf 生成コードが `generated: strict` で自動除外され、運用負荷が低い
* Bad, because 100+ linter の取捨選択・設定チューニングに学習コストがかかる

## More Information

* golangci-lint 公式ドキュメント: https://golangci-lint.run/
* golangci-lint v2 移行ガイド: https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/
* golangci-lint GitHub Action: https://github.com/golangci/golangci-lint-action
* golangci-lint GitHub リポジトリ: https://github.com/golangci/golangci-lint
