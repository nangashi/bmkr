### タスク002: 外部サービスセットアップ

**blocked_by**: なし

#### 成果物

- [x] 外部: Neon PostgreSQL プロジェクト作成（Web コンソールで操作）および DATABASE_URL 取得
- [x] 外部: Google OAuth クライアント ID・シークレット取得（Google Cloud Console で操作）
- [x] 外部: Netlify サイト作成・GitHub リポジトリ連携（Netlify コンソールで操作）

#### 受け入れ条件

- [x] DATABASE_URL 環境変数が取得済みである（Neon PostgreSQL 接続文字列形式であること）
- [x] GOOGLE_CLIENT_ID および GOOGLE_CLIENT_SECRET が取得済みである
- [x] Netlify サイトが作成され、GitHub リポジトリと連携されている（Netlify コンソールでデプロイ設定が確認できる）

#### 入力

- `architecture.md` §8.1（デプロイ構成）
- `architecture.md` §8.3（コスト設計: Netlify Starter・Neon 無料枠）
- `standards.md` §3.5（確定済み環境変数一覧）
- `requirements.md` §4（制約: NFR-1 月額0円）
