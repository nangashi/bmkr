---
status: "accepted"
date: 2026-03-22
last-validated: 2026-03-22
---

# Go サービスの構造化ログライブラリの選定

## Decision Outcome

log/slog（標準ライブラリ）を採用する。外部依存ゼロでログ設計ガイドとの完全一致が得られ、ガイドのコード例（`slog.InfoContext`, `slog.WarnContext`）がそのままコピー&ペーストで動作する。`slog.Handler` インターフェースにより将来 zap/zerolog バックエンドへの移行もアプリコード変更なしで可能である。標準ライブラリとして Go の interface 設計パターンを学べる学習価値も高く、ガイドのコード例が slog 前提で書かれているためライブラリ選定とガイドの間に乖離が生じない点が決定的だった。

### Accepted Tradeoffs

- パフォーマンスが zap / zerolog に劣る。学習プロジェクトの規模では問題にならないと判断。将来パフォーマンス要件が厳しくなった場合は、Handler を zap/zerolog バックエンドに差し替えることで対応可能

### Consequences

- `log.Printf` を `slog.InfoContext` に置き換えることで、ガイドの canonical log line パターンを実装できる
- sloglint を golangci-lint v2 (ADR-0011) に統合し、フィールド命名規約を CI で自動検証できる
- Connect RPC 用ロギングインターセプタは自作が必要（20-40行程度）

## Context and Problem Statement

3つの Go バックエンドサービス（ec-site, product-mgmt, customer-mgmt）は現在 `log.Printf` / `log.Fatalf` のみでログ出力しており、構造化ログに対応していない。ログ設計ガイド（docs/guides/go-logging.md）で JSON 構造化ログと canonical log line パターンが策定済みだが、それを実現するライブラリが未選定である。この ADR では Go バックエンドサービスのログライブラリを選定する。BFF（Fastify）のログ基盤、OpenTelemetry/分散トレーシングの導入判断はスコープ外とする。

## Prerequisites

- バックエンド API は Go (Echo)、サービス間通信に Connect RPC を採用 (ADR-0001, ADR-0003)
- Go 1.21+ の log/slog が標準ライブラリとして利用可能
- ログ設計ガイド (docs/guides/go-logging.md) 策定済み — JSON 出力、canonical log line、フィールド規約が定義されている
- 学習プロジェクトのため外部依存は最小限が望ましく、3サービスで同一ライブラリを使う
- 将来の OpenTelemetry 統合を妨げない選択であること

## Decision Drivers

- 外部依存の少なさ
- ログ設計ガイドとの適合性
- 将来の拡張パス
- Go エコシステムでの学習価値

## Considered Options

| 選択肢 | 概要 |
|--------|------|
| **log/slog（採用）** | Go 標準ライブラリの構造化ログ。追加依存ゼロ |
| uber-go/zap | 高性能な構造化ログライブラリ。実務普及度が高い |
| rs/zerolog | ゼロアロケーション設計の構造化ログライブラリ |

除外: sirupsen/logrus（メンテナンスモード宣言済み）、phuslu/log（コミュニティが小さく OTel ブリッジなし）、charmbracelet/log（CLI 向け）、go-kit/log（メンテナンス停滞）

## Comparison Overview

| 判断軸 | log/slog | uber-go/zap | rs/zerolog |
|--------|----------|-------------|------------|
| 外部依存の少なさ | ◎ 追加依存ゼロ | △ +2パッケージ（zap + multierr） | △ +3パッケージ新規追加 |
| ログ設計ガイドとの適合性 | ◎ ガイドが slog 前提、コード例がそのまま動く | △ ネイティブ API はガイドと不一致、zapslog(experimental)経由なら互換 | ○ slog Handler 経由なら互換、ネイティブ API は全面書き換え |
| 将来の拡張パス | ◎ Handler 差し替えで zap/zerolog へ移行可能、otelslog(v0.17.0) | ○ otelzap(v0.17.0)、zapcore による高い拡張性 | ○ Hook 機構で拡張可能、otelzerolog は v0.0.0 で未成熟 |
| Go エコシステムでの学習価値 | ◎ 標準 interface 設計の習得、今後のデファクト | ○ 実務普及度高、slog 時代で位置づけ変化中 | ○ ゼロアロケーション設計の学習 |

## Notes

- zapslog が experimental (v0.3.0) のまま2年以上経過し安定版昇格の見通し不明
- zap は `logger.Sync()` が必要で Docker 環境での既知バグあり (issue #1093)
- zerolog のチェイナブル API は `Msg()`/`Send()` 呼び忘れでログが消えるバグを生みやすい
- zerolog の slog Handler 経由で使うと性能差が変換レイヤーで相殺され「zerolog を入れた意味」が薄れる

## More Information

* [Go 公式ブログ: Structured Logging with slog](https://go.dev/blog/slog)
* [awesome-slog — slog エコシステムカタログ](https://github.com/go-slog/awesome-slog)
* [otelslog — OpenTelemetry 公式 slog ブリッジ](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog)
