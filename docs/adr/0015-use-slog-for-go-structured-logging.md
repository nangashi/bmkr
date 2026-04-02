---
status: "accepted"
date: 2026-03-22
last-validated: 2026-03-22
---

# Go サービスの構造化ログライブラリに log/slog を採用する

## Context and Problem Statement

3つの Go バックエンドサービス（ec-site, product-mgmt, customer-mgmt）は現在 `log.Printf` / `log.Fatalf` のみでログ出力しており、構造化ログに対応していない。ログ設計ガイド（docs/guides/go-logging.md）で JSON 構造化ログと canonical log line パターンが策定済みだが、それを実現するライブラリが未選定である。この ADR では Go バックエンドサービスのログライブラリを選定する。BFF（Fastify）のログ基盤、OpenTelemetry/分散トレーシングの導入判断はスコープ外とする。

## Prerequisites

* バックエンド API は Go (Echo) を使用する (ADR-0001)
* サービス間通信に Connect RPC を採用している (ADR-0003)
* Go 1.21+ の log/slog が標準ライブラリとして利用可能
* 現状は標準 log パッケージの Printf / Fatalf のみ使用（3サービス共通）
* ログ設計ガイド (docs/guides/go-logging.md) が策定済み — JSON 出力、canonical log line、フィールド規約が定義されている
* 学習プロジェクトのため外部依存は最小限が望ましい
* 3サービスで同一のライブラリを使う（統一性）
* 将来の OpenTelemetry 統合を妨げない選択であること

## Decision Drivers

* 外部依存の少なさ — go.mod に追加される依存パッケージ数。学習プロジェクトとしてシンプルに保てるか
* ログ設計ガイドとの適合性 — 策定済みガイド（JSON 出力、canonical log line、`slog.InfoContext` 等のコード例）をそのまま実装できるか
* 将来の拡張パス — OpenTelemetry ブリッジの公式サポート、Handler 差し替えによるバックエンド変更の容易さ
* Go エコシステムでの学習価値 — Go の標準パターン（interface ベース設計等）や実務で広く使われる知識が身につくか

## Considered Options

* log/slog（標準ライブラリ）
* uber-go/zap
* rs/zerolog

### Excluded Options

* sirupsen/logrus — 公式にメンテナンスモードを宣言済み。パフォーマンスが 2231 ns/op, 23 allocs/op と大幅に劣る。新規プロジェクトでの採用は非推奨
* phuslu/log — コミュニティが小さく学習資料が不足。OTel 公式ブリッジも存在しない
* charmbracelet/log — CLI 向けの美麗出力に特化しており、サーバーサイドのマイクロサービスには不向き
* go-kit/log — メンテナンスが停滞。slog の登場で役割が代替されている

## Comparison Overview

| 判断軸 | log/slog | uber-go/zap | rs/zerolog |
|--------|----------|-------------|------------|
| 外部依存の少なさ | ◎ 追加依存ゼロ | △ +2パッケージ（zap + multierr） | △ +3パッケージ新規追加 |
| ログ設計ガイドとの適合性 | ◎ ガイドが slog 前提、コード例がそのまま動く | △ ネイティブ API はガイドと不一致、zapslog(experimental)経由なら互換 | ○ slog Handler 経由なら互換、ネイティブ API は全面書き換え |
| 将来の拡張パス | ◎ Handler 差し替えで zap/zerolog へ移行可能、otelslog(v0.17.0) | ○ otelzap(v0.17.0)、zapcore による高い拡張性 | ○ Hook 機構で拡張可能、otelzerolog は v0.0.0 で未成熟 |
| Go エコシステムでの学習価値 | ◎ 標準 interface 設計の習得、今後のデファクト | ○ Stars 最多(24.4k)、実務普及度高、slog 時代で位置づけ変化中 | ○ ゼロアロケーション設計の学習、Stars 12.3k |

◎/○/△ は選択肢間の相対的な優劣を示す目安。

## Pros and Cons of the Options

### log/slog

* Good, because 標準ライブラリのため go.mod への依存追加がゼロ。3サービスでバージョン管理不要
* Good, because ガイドのコード例（`slog.InfoContext`, `slog.WarnContext`）がそのままコピー&ペーストで動作する
* Good, because `slog.Handler` インターフェースにより、将来 zap/zerolog バックエンドへの移行がアプリコード変更なしで可能
* Good, because `slog.Handler` の4メソッド設計は Go のインターフェース設計哲学の学習教材として優秀
* Good, because Echo v4 が `RequestLoggerWithConfig` で slog を公式サポート
* Good, because sloglint を golangci-lint v2（ADR-0011 採用済み）に統合可能
* Bad, because パフォーマンスが zap/zerolog に劣る（slog 174 ns/op vs zap 71 ns/op vs zerolog 30 ns/op）
* Bad, because Connect RPC 用ロギングインターセプタの既成品が少なく自作が必要

### uber-go/zap

* Good, because GitHub Stars 約24,400 で実務での普及度が最も高い
* Good, because zapcore による高い拡張性（カスタム Core/Encoder の組み合わせが自在）
* Good, because ベンチマーク 71 ns/op と高性能
* Bad, because ガイドのコード例（`slog.InfoContext`）と zap ネイティブ API が不一致。ガイド改訂か zapslog 経由が必要
* Bad, because zapslog が experimental (v0.3.0) のまま2年以上経過し安定版昇格の見通し不明
* Bad, because `logger.Sync()` が必要で Docker 環境での既知バグあり (issue #1093)
* Bad, because 学習プロジェクトに対して不要な外部依存を追加する

### rs/zerolog

* Good, because ベンチマーク最速（30 ns/op、ゼロアロケーション）
* Good, because デフォルト JSON 出力でガイドの「本番は JSON」方針と一致
* Good, because `zerolog.NewSlogHandler()` がネイティブ実装済みで slog API との互換性あり
* Good, because ゼロアロケーション設計は Go のメモリ管理・パフォーマンス最適化の学習素材として価値がある
* Bad, because slog Handler 経由で使うと「zerolog を入れた意味」が薄れる（性能差も変換レイヤーで相殺）
* Bad, because otelzerolog が v0.0.0 で OTel ブリッジの成熟度が最も低い
* Bad, because チェイナブル API は `Msg()`/`Send()` 呼び忘れでログが消えるバグを生みやすい
* Bad, because 外部依存が3パッケージ新規追加

## Decision Outcome

Chosen option: "log/slog"

### Rationale

外部依存ゼロ・ログ設計ガイドとの完全一致・将来の拡張パス（Handler 差し替え）の3軸を重視した。標準ライブラリとして Go の interface 設計パターンを学べる学習価値も高い。ガイドのコード例が slog 前提で書かれているため、ライブラリ選定とガイドの間に乖離が生じない点が決定的だった。

### Accepted Tradeoffs

* パフォーマンスが zap (71 ns/op) / zerolog (30 ns/op) に劣る (174 ns/op)。学習プロジェクトの規模では問題にならないと判断。将来パフォーマンス要件が厳しくなった場合は、Handler を zap/zerolog バックエンドに差し替えることで対応可能

### Consequences

* Good, because `log.Printf` を `slog.InfoContext` に置き換えることで、ガイドの canonical log line パターンを実装できる
* Good, because sloglint を golangci-lint v2 (ADR-0011) に統合し、フィールド命名規約を CI で自動検証できる
* Bad, because Connect RPC 用ロギングインターセプタは自作が必要（20-40行程度）

## More Information

* [Go 公式ブログ: Structured Logging with slog](https://go.dev/blog/slog)
* [awesome-slog — slog エコシステムカタログ](https://github.com/go-slog/awesome-slog)
* [otelslog — OpenTelemetry 公式 slog ブリッジ](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog)

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-22 | 初版作成 | N/A |
