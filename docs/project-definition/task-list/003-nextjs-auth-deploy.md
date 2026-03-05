### タスク003: Next.jsプロジェクト初期化 + Auth.js認証設定 + Netlifyデプロイ

**blocked_by**: Task 001（ツールチェーンセットアップ）, Task 002（外部サービスセットアップ）

#### 成果物

- [x] Next.js App Router プロジェクト初期化（pnpm、src/ ディレクトリ構成、app/(auth)/・app/(app)/ ルートグループ）
- [x] Auth.js + Google OAuth 認証設定（callbacks.signIn での ALLOWED_EMAIL 照合によるシングルユーザー制限・events.signInFailure ログ出力）
- [x] AuthGuard（middleware.ts: Auth.js Middleware による全ページ認証チェック、/api/ping の認証対象外設定）
- [x] Netlify デプロイ設定（netlify.toml またはビルド設定）
- [x] プロジェクトドキュメント（README.md: セットアップ手順・開発コマンド一覧・pg_dump バックアップ手順記載）
- [x] CLAUDE.md（AI コーディング支援用コンテキスト）
- [x] .env.example（環境変数キー一覧: DATABASE_URL, ALLOWED_EMAIL, GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, AUTH_SECRET, PING_SECRET）
- [x] public/robots.txt（全クローラー拒否設定）

#### 受け入れ条件

- [ ] Netlify 上にデプロイされたアプリへのアクセスで、Google OAuth ログイン画面にリダイレクトされる
- [ ] ALLOWED_EMAIL に設定したメールアドレスの Google アカウントでログインできる
- [ ] ALLOWED_EMAIL と異なるメールアドレスのアカウントでは認証が拒否される
- [ ] 未認証状態でアプリ画面へのアクセスを試みると、ログインページへリダイレクトされる
- [x] `pnpm biome check` がエラーなしで通過する
- [x] .env.example に 6 つの環境変数キーが全て記載されている

#### 入力

- `architecture.md` §7.1（認証フロー）
- `architecture.md` §7.2（アクセス制御方針・middleware 除外パス）
- `architecture.md` §8.1（デプロイ構成）
- `requirements.md` §8（NFR-1, NFR-2, NFR-4）
- `requirements.md` §10（IR-003）
- `standards.md` §3.2（ディレクトリ構成）
- `standards.md` §3.5（.env 戦略・確定済み環境変数一覧）
- `development-process.md` §1.5（ドキュメント管理）
- `development-process.md` §5.4（ドキュメント戦略）
