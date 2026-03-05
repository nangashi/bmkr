# bmkr

クロスデバイス対応リーディングリスト管理ツール。

## 技術スタック

- Next.js (App Router) + TypeScript
- Auth.js (next-auth v5 beta) + Google OAuth
- Neon PostgreSQL + Drizzle ORM (Task 004 で導入)
- Biome (linter/formatter)
- Netlify (hosting)

## ディレクトリ構成

```
src/
  app/                  # Next.js App Router
    (auth)/             # 認証関連ページ
    (app)/              # 認証済みアプリページ
    api/                # API routes
  lib/                  # ビジネスロジック・DB・型定義
    services/           # Service 層
    repositories/       # Repository 層
    interfaces/         # DI 用インターフェース
    db/                 # Drizzle ORM スキーマ
  components/           # 再利用可能 UI コンポーネント
  middleware.ts         # Auth.js Middleware
  auth.ts               # Auth.js 設定
```

## コーディングルール

- Biome recommended + noVar/useConst/noDangerouslySetInnerHtml
- TypeScript strict + noUncheckedIndexedAccess
- ファイル名: kebab-case、クラス/コンポーネント: PascalCase
- コメントは「なぜ (Why)」のみ。JSDoc 不使用
- Barrel exports (index.ts) 不使用
- 依存方向: app/ -> lib/services/ -> lib/repositories/

## 開発コマンド

```bash
pnpm dev          # 開発サーバー
pnpm build        # ビルド
pnpm lint         # Biome check
pnpm lint:fix     # Biome auto-fix
```

## 設計文書

- `docs/project-definition/` 配下に要件・アーキテクチャ・規約・開発プロセス・詳細設計
- `docs/adr/` に ADR (Architecture Decision Records)

## Development Environment

Docker container with AI development tools (Claude Code, Codex CLI, TAKT).

### Setup

1. Configure 1Password CLI references in `.mise.toml` `[env]` section to match your vault structure
2. `mise run build` — Build the container
3. `mise run up` — Start the container (1Password resolves API keys)
4. `mise run claude` / `mise run codex` / `mise run takt` — Launch AI tools

### Available Tasks

- `mise run build` — Build dev container
- `mise run up` — Start dev container (resolves 1Password secrets)
- `mise run down` — Stop dev container
- `mise run shell` — Open bash shell
- `mise run claude` — Start Claude Code
- `mise run codex` — Start Codex CLI
- `mise run takt` — Start TAKT

# currentDate
Today's date is 2026-03-05.
