---
status: "accepted"
date: 2026-04-03
last-validated: 2026-04-03
---

# サービス間共通コードの管理方針

## Decision Outcome

Go Workspaces (go.work) を採用する。IDE 体験と将来の拡張性を重視した。gopls が go.work をネイティブサポートしており、クロスモジュールの定義ジャンプ・補完が最も安定している。go.work の use ディレクティブで依存解決を一元管理でき、各 go.mod の replace ディレクティブが不要になる。既存パターンとの一貫性（gen/go の replace 運用）よりも、ツールチェーンとの統合品質と新サービス追加時の手軽さを優先した。

### Accepted Tradeoffs

- gen/go の既存 replace ディレクティブを go.work に移行する作業が発生する
- go.work をコミットする運用を前提とし、将来 Docker 化する際にビルドコンテキストへの go.work 包含または `GOWORK=off` の設定が必要になる

### Consequences

- 各 go.mod から replace ディレクティブを除去でき、go.mod がクリーンになる
- 新サービス追加時は go.work に `use ./services/new-service` を1行追加するだけで済む
- 共有モジュール（例: `lib/go`）を go.work に追加すれば、インフラ層コードの共有が replace なしで実現できる

## Context and Problem Statement

ADR-0016 で3サービス（product-mgmt / ec-site / customer-mgmt）に同一のログ出力パターンを適用することが決まったが、Connect インターセプタや Echo ミドルウェア設定など、サービス間で重複するインフラ層コードの共有方法が未定義である。現状は各サービスに約90行の同一コードがコピーとして存在し、変更時の同期が手動に依存している。この ADR ではインフラ層の共通コードの管理方針を決定する。gen/go（protobuf 生成コード）の共有方法は対象外。

## Prerequisites

- 3サービスで同一の実装パターンを使う (ADR-0016)
- gen/go が replace ディレクティブで共有モジュールとして運用されている

## Decision Drivers

- 変更同期の確実性
- 既存パターンとの一貫性
- IDE・ツールチェーン互換性
- 導入・移行コスト

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| 共有モジュール + replace | リポジトリルートに共有モジュールを作り、各サービスの go.mod で replace する。gen/go と同一パターン |
| **Go Workspaces (go.work)（採用）** | go.work の use ディレクティブで複数モジュールを一元管理。replace 不要 |
| go generate コピー | テンプレートからコードを生成し各サービスに配置。CI で差分検知 |
| ローカルパッケージ | 現状維持。各サービス内に同一コードを個別保持 |

除外: なし

## Comparison Overview

| 判断軸 | 共有モジュール + replace | Go Workspaces (go.work) | go generate コピー | ローカルパッケージ |
|--------|:-:|:-:|:-:|:-:|
| 変更同期の確実性 | ◎ ローカルパス参照で即反映、タグ付け不要 | ◎ use で列挙されたモジュールは常に最新を参照、反映漏れが構造的に起きない | ○ CI 差分検知で防止可能だが generate 実行忘れのリスク | △ 手動コピーが必要、ガードレール構築が別途必要 |
| 既存パターンとの一貫性 | ◎ gen/go と完全に同一パターン、追加学習コストなし | ○ gen/go の replace を go.work に置き換え可能だが移行が必要 | △ gen/go とは異なる仕組みが並存、理解コスト増 | △ gen/go が replace で共有されている前例と矛盾 |
| IDE・ツールチェーン互換性 | ○ gopls は replace を解決可能、ただしマルチモジュール認識に設定が要る場合あり | ◎ gopls が go.work をネイティブサポート、クロスモジュールのジャンプ・補完が最も安定 | ◎ 生成ファイルは通常の .go なので完全に動作 | ◎ 各サービス自己完結で問題なし |
| 導入・移行コスト | ◎ go.mod に replace 1行追加 + import パス変更、gen/go と同じ手順 | ◎ go work init 1コマンド + 各 go.mod の replace 削除 | △ テンプレート + ジェネレータ + CI 差分検知の構築が必要 | ◎ 現状維持でゼロコスト |

## Notes

- **`internal/` は共有モジュール名に使えない**: Go の internal パッケージ可視性ルールにより別モジュールからのインポートがコンパイラレベルで拒否される。共有モジュールには `lib/go` 等の名前が必要
- **golangci-lint の go.work 対応**: go.work をコミットする場合、golangci-lint が go.work を正しく扱うかは別途検証が必要

## More Information

- [Tutorial: Getting started with multi-module workspaces](https://go.dev/doc/tutorial/workspaces)
- [Get familiar with workspaces - Go Blog](https://go.dev/blog/get-familiar-with-workspaces)
