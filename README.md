# bmkr

クロスデバイス対応リーディングリスト管理ツール。

## セットアップ

### 前提条件

- Node.js 22
- pnpm

### インストール

```bash
pnpm install
```

### 環境変数

`.env.example` をコピーして `.env.local` を作成し、値を設定する。

```bash
cp .env.example .env.local
```

| 変数名 | 用途 |
|--------|------|
| `DATABASE_URL` | Neon PostgreSQL 接続文字列 |
| `ALLOWED_EMAIL` | シングルユーザー許可メールアドレス |
| `GOOGLE_CLIENT_ID` | Google OAuth クライアント ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth クライアントシークレット |
| `AUTH_SECRET` | Auth.js セッション署名用シークレット (`npx auth secret` で生成) |
| `PING_SECRET` | warm-up ping エンドポイントの認証トークン |

## 開発コマンド

```bash
pnpm dev          # 開発サーバー起動
pnpm build        # プロダクションビルド
pnpm start        # プロダクションサーバー起動
pnpm lint         # Biome lint/format チェック
pnpm lint:fix     # Biome 自動修正
```

## デプロイ

main ブランチへの push で Netlify が自動デプロイを実行する。

## バックアップ

Neon PostgreSQL の週次バックアップ手順:

```bash
pg_dump "$DATABASE_URL" --no-owner --no-acl > backup_$(date +%Y%m%d).sql
```

Neon のポイントインタイムリストアも有効化済み。
