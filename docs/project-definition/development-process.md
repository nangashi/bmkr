# クロスデバイス対応リーディングリスト管理ツール: 開発プロセス設計書

## ステータス

Confirmed

## 日付

2026-02-27

## 入力文書

- アーキテクチャ設計書: `docs/project-definition/architecture.md` (2026-02-26)
- 開発規約書: `docs/project-definition/standards.md` (2026-02-27)
- 要件定義書: `docs/project-definition/requirements.md` (2026-02-19)
- ADR: `docs/adr/` (5件: ADR-0001, ADR-0002, ADR-0003, ADR-0004, ADR-0005)

## 1. Git ワークフロー・開発プロセス

### 1.1 Git ワークフロー（ADR 参照）

**Trunk-based Development（簡略版）**を採用する（設計決定 G1-1）。

- 初期開発中: main ブランチに直接 push
- 初期開発完了後: feature ブランチ経由のマージに移行（main への直接 push を禁止）
- 複数日にわたる変更や実験的な変更は短命な feature ブランチを作成（生存期間: 原則 1〜3 日以内）

Netlify のデプロイが main ブランチ push に連動しており、個人開発の規模感に適合する。

### 1.2 ブランチ戦略

**P1/P2 導出根拠**: 設計決定 G1-2（G1-1 から導出）

- **長命ブランチ**: `main` のみ
- **短命ブランチ命名規則**: `{type}/{description}`（例: `feat/article-save`, `fix/title-fetch-timeout`）。type は Conventional Commits（standards.md §1.6）と統一する
- **ブランチ保護**:
  - 初期開発中: なし（main への直接 push を許可）
  - 初期開発完了後: GitHub ブランチ保護ルールで main への直接 push を禁止
- **マージ戦略**: feature ブランチ使用時は squash merge（1 ブランチ = 1 commit）

### 1.3 Issue 管理（ADR 参照）

**GitHub Issues** を採用する（設計決定 G1-3）。

- Issue テンプレートは設けない（個人開発のため過剰）
- ラベルは必要に応じて追加（`bug`, `feat`, `chore` 程度）
- コミットメッセージから `#123` で自動リンク

根拠: GitHub 統一・追加コスト不要。

### 1.4 リリースプロセス

**P1/P2 導出根拠**: 設計決定 G1-4（G1-1 から導出）、NFR-2（1コマンド以内）

main push = 自動リリース。バージョン管理は導入しない。

- `git push origin main` で Netlify が自動ビルド・デプロイ
- マイグレーションを含むデプロイの場合のみ、手動で `pnpm drizzle-kit migrate` を先行実行してからプッシュ
- ロールバックは Netlify のインスタントロールバック機能を利用
- CHANGELOG 自動生成は不要

### 1.5 ドキュメント管理

**P1/P2 導出根拠**: 設計決定 G1-5

**Docs-as-Code**（リポジトリ内 Markdown 管理）を採用する。

- **ADR**: `docs/adr/` に連番ファイルで管理（既存方式を継続）
- **設計文書**: `docs/project-definition/` に配置
- **README.md**: プロジェクトルートにセットアップ手順・概要・開発コマンド一覧を記載
- **CLAUDE.md**: AI コーディング支援用コンテキスト管理
- **.env.example**: 環境変数キー一覧（値は空またはダミー）
- 作成しないもの: CONTRIBUTING.md、API ドキュメント、CHANGELOG.md
- 外部ドキュメントツールは使用しない

## 2. テスト戦略

### 2.1 テスト投資配分

**P1/P2 導出根拠**: 設計決定 G2-1、architecture.md §5.3

**テスティングトロフィー**（結合テスト重視）を採用する。

| レベル | 責務 | 対象 | 実行タイミング |
|--------|------|------|--------------|
| Static Analysis | 型エラー・構文・スタイルの検出 | 全コード | 保存時（エディタ）/ CI Stage 1 |
| Unit | ビジネスロジックの検証（外部依存はモック差し替え） | Service 層（ArticleService, TitleFetcher） | CI Stage 1 |
| Integration（重点） | Drizzle ORM クエリと Neon PostgreSQL の結合検証。Server Actions の入出力検証 | Repository 層 + 実 DB | CI Stage 2 |
| E2E（主要シナリオのみ） | ブラウザ操作の主要フロー検証（保存・既読化・検索） | Handler 層〜DB 全体 | CI Stage 2 |

### 2.2 テストフレームワーク（ADR 参照）

**選定結果**: Vitest（Unit/Integration）+ Playwright（E2E）を採用する（ADR-0005）。

- 全評価目的（App Router 対応能力、TypeScript strict 統合品質、開発フィードバック速度、E2E ブラウザ対応柔軟性、導入・維持コストの低さ）で最優位。トレードオフなし
- 非同期 RSC のユニットテスト不可は全フレームワーク共通の制約。E2E に委ねる

### 2.3 テストカバレッジ目標

**P1/P2 導出根拠**: 設計決定 G2-3

数値目標は設定しない。

- 定性的基準: 主要ユースケース（UC-1〜UC-8）のパスがテストで網羅されていること
- カバレッジ計測ツール（Vitest の `--coverage` オプション）は導入するが、カバレッジゲートは設けない（参考値として出力）

### 2.4 テストファイル配置

テストファイルはテスト対象ファイルと同一ディレクトリに colocate する。

```
src/lib/services/
  article-service.ts
  article-service.test.ts
  title-fetcher.ts
  title-fetcher.test.ts
src/lib/repositories/
  article-repository.ts
  article-repository.test.ts
```

E2E テストは `e2e/` ディレクトリにまとめて配置する。

```
e2e/
  article-save.spec.ts
  article-read.spec.ts
  search.spec.ts
```

### 2.5 テストデータ管理

**P1/P2 導出根拠**: 設計決定 G2-4

**インラインファクトリー関数方式**を採用する。

- テストファイル内にヘルパー関数（`buildArticle()` 等）を定義
- 結合テスト用 DB シードは `beforeEach` 内で直接 INSERT
- 外部フィクスチャファイルは使用しない

### 2.6 モック戦略

**P1/P2 導出根拠**: 設計決定 G2-5、architecture.md §5.3（DI 設計）

**DI ベースのモック注入**を基本とする。MSW は導入しない。

- architecture.md §5.3 の DI 設計（コンストラクタ注入）を活用する
- `ITitleFetcher`、`IArticleRepository` のモック実装をコンストラクタ注入で差し替え
- 結合テストは実 DB（Neon 開発用ブランチ）を使用する

### 2.7 スナップショットテストの採否

**不採用**（設計決定 G2-6）。

少数 UI + shadcn/ui 構成で、スナップショットの維持コストに見合う価値がない。

### 2.8 CI でのテスト実行戦略

**P1/P2 導出根拠**: 設計決定 G2-7

2 ステージ構成で全テストを実行する。

| ステージ | 内容 | 実行タイミング |
|---------|------|-------------|
| Stage 1: Fast（並列実行） | Biome lint/format check + TypeScript type check（`tsc --noEmit`）+ Unit tests（Vitest）+ gitleaks | 全 push 時 |
| Stage 2: Heavy（Stage 1 成功後） | Integration tests（Vitest + 実 DB）+ E2E tests（Playwright）+ Build verification | 全 push 時（初期開発後は main マージ時のみに変更検討可） |

## 3. CI/CD・環境管理

### 3.1 CI/CD プラットフォーム（ADR 参照）

**選定結果**: GitHub Actions（CI）+ Netlify ビルトイン CD（デプロイ）の組み合わせを採用する（設計決定 G3-1、ADR-0001）。

| 役割 | ツール | 担当範囲 |
|------|--------|---------|
| CI（品質ゲート） | GitHub Actions | lint, 型チェック, テスト, シークレットスキャン, `pnpm audit` |
| CD（ビルド・デプロイ） | Netlify ビルトイン | ビルド + 自動デプロイ（GitHub Actions とは独立） |

### 3.2 CI パイプライン構成

**P1/P2 導出根拠**: 設計決定 G3-2

| ステージ | ジョブ | トリガー | 備考 |
|---------|--------|---------|------|
| Stage 1: Fast | Biome lint check | 全 push | `pnpm biome check --reporter=github` |
| Stage 1: Fast | TypeScript type check | 全 push | `tsc --noEmit` |
| Stage 1: Fast | Unit tests | 全 push | `pnpm vitest run --reporter=verbose` |
| Stage 1: Fast | gitleaks（シークレットスキャン） | 全 push | GitHub Actions で実行 |
| Stage 1: Fast | pnpm audit | 全 push | 警告出力のみ（ゲートブロックなし） |
| Stage 2: Heavy | Integration tests | Stage 1 成功後・全 push | `pnpm vitest run --reporter=verbose`（DB 接続要） |
| Stage 2: Heavy | E2E tests | Stage 1 成功後・全 push | `pnpm playwright test` |
| Stage 2: Heavy | Build verification | Stage 1 成功後・全 push | `pnpm build` |

### 3.3 CD 戦略

**P1/P2 導出根拠**: 設計決定 G3-3、NFR-2

- **デプロイフロー**: main push → Netlify が自動ビルド・デプロイ（手動承認ステップなし）
- **デプロイプレビュー**: feature ブランチ PR 時に Netlify Deploy Preview を利用（初期開発完了後）
- **ロールバック**: Netlify インスタントロールバック機能を利用
- **マイグレーション含むデプロイ**: スキーマ変更時のみ手動で `pnpm drizzle-kit migrate`（本番 DB）を先行実行してからプッシュ

### 3.4 IaC の採否（ADR 参照）

**該当なし**（設計決定 G3-4）。

管理対象が Netlify + Neon の 2 サービスのみであり、IaC の導入・維持コストに見合わないと判断した。

### 3.5 環境管理

**P1/P2 導出根拠**: 設計決定 G3-5、architecture.md §8.2

| 環境 | 用途 | デプロイ条件 | 備考 |
|------|------|------------|------|
| 本番 | 実運用 | main ブランチへの push 時に Netlify が自動デプロイ | Netlify main デプロイ + Neon 本番 DB |
| 開発 | ローカル開発・テスト | 手動起動（`pnpm dev`） | next dev + Neon 開発用ブランチ（またはローカル PostgreSQL） |

ステージング環境は設けない。個人開発・シングルユーザー構成で維持コストに見合わない。

### 3.6 シークレット管理（CI/CD 内）

**P1/P2 導出根拠**: 設計決定 G3-6、設計決定 G4-4

**GitHub Actions Secrets + Netlify 環境変数**の二系統で管理する。

| 環境 | 管理方式 |
|------|---------|
| ローカル開発 | `.env.local`（Git 除外、`.gitignore` で管理） |
| CI（GitHub Actions） | GitHub Actions Secrets |
| 本番（Netlify） | Netlify 環境変数（Netlify コンソールで設定） |

シークレットのローテーション: 定期ローテーションなし。漏洩疑い時に手動更新。

## 4. インフラ・運用プロセス

### 4.1 マイグレーション戦略

**P1/P2 導出根拠**: 設計決定 G4-1、architecture.md §8.1

**Drizzle Kit マイグレーション（手動実行・デプロイ前）**を採用する。

ワークフロー:
1. `src/lib/db/schema.ts` を変更する
2. `pnpm drizzle-kit generate` → マイグレーション SQL を生成する
3. 生成された SQL を目視確認する
4. 開発 DB で `pnpm drizzle-kit migrate` → 動作確認する
5. 本番 DB で `pnpm drizzle-kit migrate` を先行実行する
6. `git push origin main` → Netlify 自動デプロイ

ロールバック: 手動で逆方向の SQL を作成・適用する（失敗時）。

### 4.2 シードデータ管理

**該当なし**（設計決定 G4-2）。

シードスクリプトは作成しない。テストデータは結合テストの `beforeEach` 内で直接 INSERT する方式を採用する（§2.4 参照）。

### 4.3 データバックアップ・リストア

**P1/P2 導出根拠**: 設計決定 G4-3、architecture.md §9.4

| 項目 | 内容 |
|------|------|
| 主バックアップ | Neon ポイントインタイムリストア（PITR）— 保持期間約 24 時間（無料プラン制約） |
| 補完バックアップ | `pg_dump` 週次手動実行: `pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql` |
| 保存先 | ローカルマシン（初期運用時） |
| RPO | 直近バックアップ時点（最大 1 週間） |
| RTO | 1 日以内（Neon Console または CLI からリストア） |
| 稼働率 SLA | 設定しない（個人利用・SLA 不要） |

**重要**: 初回本番デプロイ前に `pg_dump` バックアップ手順を README に記載すること。

### 4.4 シークレット管理（アプリケーション）

**P1/P2 導出根拠**: 設計決定 G4-4、NFR-4（ソースコードに認証情報を含めない）

| 変数名 | 用途 | 管理場所 |
|--------|------|---------|
| `DATABASE_URL` | Neon PostgreSQL 接続文字列 | `.env.local` / Netlify 環境変数 |
| `ALLOWED_EMAIL` | シングルユーザー許可メールアドレス | `.env.local` / Netlify 環境変数 |
| `GOOGLE_CLIENT_ID` | Google OAuth クライアント ID | `.env.local` / Netlify 環境変数 |
| `GOOGLE_CLIENT_SECRET` | Google OAuth クライアントシークレット | `.env.local` / Netlify 環境変数 |
| `AUTH_SECRET` | Auth.js セッション署名用シークレット | `.env.local` / Netlify 環境変数 |
| `PING_SECRET` | warm-up ping エンドポイントの認証トークン | `.env.local` / Netlify 環境変数 |

`.env.example` をリポジトリに含め、変数名一覧（値は空）を管理する。`.env.local` は `.gitignore` で除外する。

### 4.5 依存関係の脆弱性管理

**P1/P2 導出根拠**: 設計決定 G4-5、standards.md §9 P3 引き継ぎ

- **`pnpm audit`**: CI Stage 1 に組み込む（警告出力のみ、ゲートブロックなし）
- **GitHub Dependabot alerts**: 脆弱性の通知・PR 自動作成（`.github/dependabot.yml` で設定）
- **gitleaks**: ソースコード内シークレットスキャン
  - lefthook pre-commit フックで commit 時に実行
  - CI Stage 1 でも実行（最終保証）
- **lefthook pre-commit フック構成**:
  - Biome check（lint/format チェック、staged files 対象、自動修正なし）
  - gitleaks（シークレットスキャン）

`.github/dependabot.yml` 設定:
```yaml
version: 2
updates:
  - package-ecosystem: npm
    directory: /
    schedule:
      interval: weekly
```

Drizzle の major バージョンアップは PR 内容を慎重に確認してからマージすること（ADR-0003 のトレードオフとして受け入れ済み）。

### 4.6 エラートラッキング（ADR 参照）

**選定結果**: MVP では **Netlify ビルトインログ + `console.error`** で対応する（設計決定 G4-6）。

設定方針:
- システムエラー: `console.error` で出力（Netlify のビルトインログで捕捉）
- 業務エラー（重複 URL 保存、TitleFetcher 失敗）: `console.warn` で出力
- 認証失敗: Auth.js `events.signInFailure` コールバックで `console.error` 出力

Sentry 等のエラー監視サービスは初期開発完了後に再検討する。

### 4.7 ヘルスチェック・稼働監視

**P1/P2 導出根拠**: 設計決定 G4-7、architecture.md §9.5

warm-up ping エンドポイント（`/api/ping`）を死活監視として兼用する。外部監視サービスは導入しない。

- 実装: `src/app/api/ping/route.ts` に軽量クエリ（`SELECT 1`）エンドポイントを作成
- 実行: Netlify Scheduled Functions で 5〜10 分間隔で定期実行
- 認証: `Authorization: Bearer <PING_SECRET>` ヘッダーによるトークン検証
- middleware 除外: `/api/ping` は Auth.js Middleware の認証対象外に設定

**Netlify Starter プランでの Scheduled Functions 利用不可時の代替手段**（P2 で確認後に選択）:
- GitHub Actions scheduled workflow による定期 ping
- Neon Auto-suspend の無効化
- Neon のコンピュート設定での長時間 suspend 遅延設定

### 4.8 パフォーマンス計測

**P1/P2 導出根拠**: architecture.md §9.6 P3 引き継ぎ、standards.md §9 P3 引き継ぎ

MVP では **手動計測**（ブラウザ DevTools の Lighthouse）で対応する。CI/CD へのパフォーマンス計測の組み込み（Lighthouse CI 等）は初期開発完了後に検討する。

計測対象（architecture.md §9.6 より）:
- Lighthouse スコア
- Core Web Vitals
- NFR-3 の応答時間基準（画面表示 3 秒以内、記事保存 5 秒以内、検索 3 秒以内）

### 4.9 Service Worker 設定

**P1/P2 導出根拠**: 設計決定 G4-8、architecture.md §5.1（PWA 採用）

**最小限の Service Worker 構成**を採用する（Web Share Target API の実現に必要な範囲のみ）。

- **ツール**: next-pwa（またはサービスワーカー直接実装）
- **キャッシュ戦略**: Network First（オンライン前提、オフライン動作は不要）
- **manifest.json**: `public/manifest.json` に配置
- **PWA アイコン**: `public/` ディレクトリに配置

## 5. 開発環境・AI 統合

### 5.1 デバッグ環境

**P1/P2 導出根拠**: 設計決定 G5-1

**VS Code + ブラウザ DevTools + next dev** の組み合わせを採用する。

VS Code 設定ファイル:
- `.vscode/launch.json`: Next.js サーバーサイドデバッグ用設定（Node.js アタッチ設定）
- `.vscode/settings.json`: Biome 拡張の有効化（保存時自動フォーマット・lint）
- `.vscode/extensions.json`: 推奨拡張機能リスト（Biome、Playwright Test Explorer 等）

### 5.2 AI コーディング支援の設定

**P1/P2 導出根拠**: 設計決定 G5-2

`CLAUDE.md` によるコンテキスト管理を採用する。開発の進行に合わせて内容を更新する。記載内容: 技術スタック概要、ディレクトリ構成、コーディングルール（standards.md への参照）、アーキテクチャの主要決定事項。

### 5.3 依存関係の更新戦略

**P1/P2 導出根拠**: 設計決定 G5-3

**GitHub Dependabot version updates（週次）**を採用する。

- 設定ファイル: `.github/dependabot.yml`（§4.5 と共用）
- Drizzle の major バージョンアップは PR 内容を慎重に確認してからマージ（ADR-0003 トレードオフ対応）
- セキュリティアップデートは GitHub Dependabot security alerts で自動通知・PR 作成

### 5.4 ドキュメント戦略

**P1/P2 導出根拠**: 設計決定 G5-4

最小限の構成を維持する:

| ドキュメント | 内容 | 配置 |
|------------|------|------|
| `README.md` | プロジェクト概要、セットアップ手順、開発コマンド一覧、pg_dump バックアップ手順 | リポジトリルート |
| `CLAUDE.md` | AI コーディング支援用コンテキスト | リポジトリルート |
| `.env.example` | 環境変数キー一覧（値は空またはダミー） | リポジトリルート |
| ADR | アーキテクチャ決定記録 | `docs/adr/` |
| 設計文書 | 要件・アーキテクチャ・規約・開発プロセス | `docs/project-definition/` |

作成しないもの: CONTRIBUTING.md、API ドキュメント、CHANGELOG.md

### 5.5 AI コスト管理・モニタリング（省略可能）

省略（AI 統合は MVP 後に決定）。

AI 機能（UC-11〜UC-16）は MVP スコープ外（requirements.md §5.3）。

### 5.6 プロンプトのバージョン管理（省略可能）

省略（AI 統合は MVP 後に決定）。

### 5.7 AI レスポンスの品質保証・テスト（省略可能）

省略（AI 統合は MVP 後に決定）。

## 6. 品質チェック配置サマリ

| チェック | 保存時 | commit 時 | push 時 | CI | CD | 備考 |
|---------|:-----:|:--------:|:------:|:--:|:--:|------|
| Biome lint | ○ | ○ | — | ○ | — | エディタ保存時（即時フィードバック）+ lefthook pre-commit（作業漏れ検出）+ CI Stage 1（最終保証） |
| Biome format | ○ | ○ | — | ○ | — | エディタ保存時（即時フィードバック）+ lefthook pre-commit（作業漏れ検出）+ CI Stage 1（最終保証） |
| Biome organizeImports | ○ | — | — | — | — | エディタ保存時のみ自動整理 |
| TypeScript type check | ○ | — | — | ○ | — | エディタ保存時（即時フィードバック）+ CI Stage 1（最終保証） |
| シークレットスキャン（gitleaks） | — | ○ | — | ○ | — | lefthook pre-commit + CI Stage 1 |
| Unit tests（Vitest） | — | — | — | ○ | — | CI Stage 1 |
| Integration tests（Vitest + 実 DB） | — | — | — | ○ | — | CI Stage 2 |
| E2E tests（Playwright） | — | — | — | ○ | — | CI Stage 2 |
| Build verification | — | — | — | ○ | ○ | CI Stage 2 + Netlify ビルト（独立実行） |
| pnpm audit | — | — | — | ○ | — | CI Stage 1（警告出力のみ、ゲートブロックなし） |
| セキュリティヘッダー（X-Frame-Options 等） | — | — | — | — | ○ | `next.config.js` でリクエスト時自動適用 |
| CSP ヘッダー | — | — | — | — | ○ | `next.config.js` で初期開発時に設定 |
| CSRF 保護 | — | — | — | — | ○ | Next.js ビルトイン（Server Actions）・Route Handler Origin 検証（リクエスト時自動） |
| DB 制約（UNIQUE / NOT NULL / ENUM） | — | — | — | — | ○ | Neon PostgreSQL（DB 書き込み時自動） |

設計原則:
- **commit 時**: lefthook pre-commit で gitleaks（シークレットスキャン）+ Biome check（lint/format チェック、自動修正なし）を実行する。実装時に LLM が Biome check + 修正を実施し、commit hook は作業漏れの検出を担う
- **push 時**: フックを導入しない
- **CI**: 全量チェック（Stage 1: 高速チェック、Stage 2: 重量チェック）
- **CD**: デプロイ時の自動適用チェック（セキュリティヘッダー・DB 制約）
- **重複許容**: エディタ保存時と CI の重複は目的が異なるため許容（エディタ=即時フィードバック、CI=最終保証）

## 7. P1/P2 トレーサビリティ

| P1/P2 決定 | 導出された P3 プロセス |
|-----------|----------------------|
| ADR-0001: Netlify（Starter）採用 | §3.1 CI/CD プラットフォーム（Netlify ビルトイン CD）、§3.3 CD 戦略（main push 自動デプロイ）、§3.5 環境管理（2 環境構成）、§3.6 シークレット管理 CI/CD 内（Netlify 環境変数）、§4.7 ヘルスチェック（warm-up ping: Scheduled Functions 利用可否確認）|
| ADR-0002: Next.js + Node.js + pnpm 採用 | §1.1 Git ワークフロー（Trunk-based Dev + Netlify 自動デプロイ）、§3.2 CI パイプライン（`pnpm biome check`, `tsc --noEmit`, `pnpm build`）、§5.1 デバッグ環境（VS Code + next dev）|
| ADR-0003: Neon (PostgreSQL) + Drizzle ORM 採用 | §4.1 マイグレーション戦略（Drizzle Kit 手動実行ワークフロー）、§4.3 データバックアップ（Neon PITR + pg_dump 週次）、§4.7 ヘルスチェック（warm-up ping でコールドスタート対策）、§5.3 依存関係更新戦略（Drizzle major 更新の慎重確認）|
| ADR-0004: Biome 単独採用 | §2.7 CI テスト実行戦略（Stage 1: Biome lint/format check）、§6 品質チェック配置サマリ（Biome lint/format/organizeImports の配置）|
| ADR-0005: Vitest + Playwright 採用 | §2.1 テスト投資配分（テスティングトロフィー）、§2.2 テストフレームワーク、§2.7 CI テスト実行戦略（Stage 1: Vitest Unit、Stage 2: Vitest Integration + Playwright E2E）|
| architecture.md §5.3: DI 設計（コンストラクタ注入）確定 | §2.5 モック戦略（DI ベースモック注入、ITitleFetcher / IArticleRepository の差し替え）|
| architecture.md §8.1: デプロイ構成（Netlify + GitHub）確定 | §1.4 リリースプロセス（git push 1コマンド自動デプロイ）、§3.3 CD 戦略（手動承認なし自動デプロイ）|
| architecture.md §9.1: エラーハンドリング方針確定 | §4.6 エラートラッキング（Netlify ビルトインログ + console.error/warn 方針）|
| architecture.md §9.4: バックアップ・リカバリ方針確定 | §4.3 データバックアップ・リストア（Neon PITR + pg_dump 週次、RPO/RTO の具体値）|
| architecture.md §9.5: Neon コールドスタート対策確定 | §4.7 ヘルスチェック・稼働監視（warm-up ping 実装・Scheduled Functions 設定・PING_SECRET 認証）|
| architecture.md §9.6 P3 引き継ぎ: セキュリティ自動化 | §4.5 依存関係の脆弱性管理（gitleaks + pnpm audit + Dependabot）|
| standards.md §8: 自動強制サマリ確定 | §6 品質チェック配置サマリ（保存時/commit 時/CI の配置と重複排除方針）|
| standards.md §9 P3 引き継ぎ: マイグレーション運用 | §4.1 マイグレーション戦略（Drizzle Kit 手動ワークフロー詳細）|
| standards.md §9 P3 引き継ぎ: warm-up ping 運用 | §4.7 ヘルスチェック（Scheduled Functions 利用可否・代替手段選択）|
| architecture.md §9.6 P3 引き継ぎ: パフォーマンス計測 | §4.8 パフォーマンス計測（MVP 手動計測、CI 組み込みは初期開発後に検討）|
| requirements.md NFR-2: リリース 1 コマンド以内 | §1.4 リリースプロセス（git push 1コマンド自動デプロイ）|
| requirements.md §5.3: AI 機能 MVP スコープ外 | §5.5〜§5.7 AI 統合（全て省略）|
