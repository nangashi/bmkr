# クロスデバイス対応リーディングリスト管理ツール: 開発規約書

## ステータス

Confirmed

## 日付

2026-02-27

## 入力文書

- アーキテクチャ設計書: `docs/project-definition/architecture.md` (2026-02-26)
- 要件定義書: `docs/project-definition/requirements.md` (2026-02-19)
- ADR: `docs/adr/` (4件: ADR-0001, ADR-0002, ADR-0003, ADR-0004)

## 1. コーディングツール・スタイル

### 1.1 Linter / Formatter（ADR-0004）

**選定結果**: Biome 単独（リント+フォーマット統一）を採用する（ADR-0004）。

**設定方針**:
- 設定ファイルは `biome.json` 1ファイルに集約する
- インポート自動ソート（`organizeImports`）を有効化する
- 推奨ルールセット（`recommended: true`）を基準とする
- `noVar`、`useConst`、`noDangerouslySetInnerHtml` ルールを有効化する
- ESLint系・Prettier は導入しない

**受け入れたトレードオフ**（ADR-0004）:
- `eslint-config-next` 相当の Next.js 固有ルールセットは Biome に存在しない
- `typescript-eslint` の型考慮ルール（型情報を利用した深い検証）は Biome に存在しない
- カスタム ESLint プラグインが将来必要になった場合は ESLint の部分的追加を検討する

### 1.2 TypeScript 設定

**P1 導出根拠**: ADR-0002（TypeScript 必須制約）、設計決定 G1-3

`tsconfig.json` に以下を設定する:

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true
  }
}
```

- `strict: true`: 型安全性を最大化する（`strictNullChecks`、`strictFunctionTypes` 等を含む）
- `noUncheckedIndexedAccess: true`: 配列・オブジェクトのインデックスアクセス結果を `T | undefined` として扱い、未定義アクセスによるバグを防止する
- Next.js デフォルトの `tsconfig.json` をベースとして上記を追加設定する

### 1.3 命名規則

**P1 導出根拠**: architecture.md §5.4 確定、設計決定 G1-4

| 対象 | ルール | 例 |
|------|--------|-----|
| ファイル名 | kebab-case | `article-service.ts`, `article-card.tsx` |
| TypeScript クラス名 | PascalCase | `ArticleService`, `ArticleRepository` |
| TypeScript インターフェース名 | PascalCase | `IArticleRepository` |
| DI インターフェース型名 | `I{ClassName}` | `ITitleFetcher`, `IArticleRepository` |
| 環境変数 | UPPER_SNAKE_CASE | `DATABASE_URL`, `ALLOWED_EMAIL`, `PING_SECRET` |
| React コンポーネント | PascalCase | `ArticleCard`, `ArticleListPage` |
| 関数・変数 | camelCase | `saveArticle`, `articleList` |
| 定数（モジュールレベル） | UPPER_SNAKE_CASE | `MAX_UNREAD_WARNING = 20` |
| 型・型エイリアス | PascalCase | `ActionResult`, `AppError` |
| Zod スキーマ変数 | camelCase + `Schema` サフィックス | `articleUrlSchema` |
| DB カラム名 | snake_case | `saved_at`, `read_at` |
| DB テーブル名 | 複数形 snake_case | `articles` |

### 1.4 コメント・ドキュメント方針

**P1 導出根拠**: 設計決定 G1-5（個人開発 + TypeScript による自己文書化）

- コメントは「なぜ（Why）」の説明に限定する
- 「何を（What）」コメントは原則不要（コードが自己説明的であること）
- JSDoc / TSDoc は使用しない
- ビジネスルールの背景や設計上の判断（ADR 参照など）を記述する場合は行コメントで記載する

### 1.5 インポート順序

**P1 導出根拠**: 設計決定 G1-6（Biome 選定で確定）

Biome の `organizeImports` による自動ソートを使用する。順序の基準:
1. Node.js 組み込みモジュール（`node:fs` 等）
2. 外部パッケージ（`react`, `next`, `drizzle-orm` 等）
3. `@/` プレフィックスによる内部パス（`@/lib/services/...` 等）
4. 相対パス（`./`, `../`）

保存時に Biome が自動整理するため手動操作は不要。

### 1.6 Git commit メッセージ規約

**P1 導出根拠**: 設計決定 G1-7

Conventional Commits 形式を採用する:

```
<type>(<scope>): <description>
```

| type | 用途 |
|------|------|
| `feat` | 新機能の追加 |
| `fix` | バグ修正 |
| `refactor` | 機能変更を伴わないリファクタリング |
| `docs` | ドキュメントのみの変更 |
| `chore` | ビルド設定・依存パッケージ管理等 |
| `test` | テストコードの追加・変更 |

- `commitlint` は導入しない（個人開発のため自動強制は過剰）
- 手動遵守とする

## 2. コードパターン

### 2.1 エラーハンドリング（P1 方針からの導出を明示）

**P1 導出根拠**: architecture.md §9.1 のコードパターン化、設計決定 G2-1

**P1 の全体方針**: バリデーション / 業務 / システムの 3 分類でエラーを処理する。

**エラー型定義**（`src/lib/errors.ts`）:

```typescript
type AppError = {
  kind: 'validation' | 'duplicate' | 'title_fetch_failed' | 'system';
  message: string;
}
```

**Server Actions 返却型**（`src/lib/types.ts`）:

```typescript
type ActionResult<T = void> = { success: true; data: T } | { success: false; error: string }
```

**各層のエラー処理パターン**:

| 層 | パターン |
|----|---------|
| Handler 層（Server Actions / Route Handlers） | `try-catch` で `AppError` を捕捉し `ActionResult` に変換して `return` する。`throw` しない |
| Service 層 | ビジネスルール違反時に `AppError` を `throw` する |
| Repository 層 | DB 固有エラー（PostgreSQL `23505` 等）を捕捉し `AppError` に変換して `throw` する |

**Error Boundary 配置**:
- `app/error.tsx`: クライアントコンポーネントの未処理エラーを補足
- `app/global-error.tsx`: 最上位エラー境界

**Server Actions でのエラー返却例**:
```typescript
async function saveArticle(url: string): Promise<ActionResult> {
  try {
    await articleService.save(url);
    return { success: true, data: undefined };
  } catch (err) {
    const appError = err as AppError;
    return { success: false, error: appError.message };
  }
}
```

### 2.2 非同期処理

**P1 導出根拠**: 設計決定 G2-2（TypeScript 親和性 + P1 TitleFetcher 設計）

- `async/await` に統一する。`.then()` / `.catch()` チェーンは使用しない
- Service / Repository 層は失敗時に `throw`（`AppError`）し、Handler 層の `try-catch` で一括捕捉する
- 独立した非同期処理は `Promise.all()` で並行実行する
- 外部通信は `AbortController` + `setTimeout` でタイムアウト制御を実装する

**TitleFetcher のタイムアウト実装パターン**（3秒）:
```typescript
const controller = new AbortController();
const timeoutId = setTimeout(() => controller.abort(), 3000);
try {
  const response = await fetch(url, { signal: controller.signal });
  // ...
} finally {
  clearTimeout(timeoutId);
}
```

### 2.3 不変性・純粋関数

**P1 導出根拠**: 設計決定 G2-3

- `const` をデフォルトとし、再代入が必要な場合のみ `let` を使用する
- `var` は禁止（Biome `noVar` ルールで自動強制）
- 関数引数のオブジェクト・配列を直接変更しない。スプレッド構文でコピーする:
  ```typescript
  // NG: 引数の直接変更
  function addTag(article: Article, tag: string): void { article.tags.push(tag); }
  // OK: スプレッドでコピー
  function addTag(article: Article, tag: string): Article { return { ...article, tags: [...article.tags, tag] }; }
  ```
- React 状態の更新は不変パターン（スプレッド / `map` / `filter`）を使用する
- Service 層のロジックは可能な限り純粋関数として実装し、副作用は Repository / TitleFetcher に集約する

### 2.4 ログ出力（形式、必須フィールド、レベル設計）

**P1 導出根拠**: architecture.md §9.6（Netlify ログ前提）、設計決定 G2-4

- ログライブラリは導入しない。`console.log` / `console.error` / `console.warn` を直接使用する
- 構造化ログは MVP では不導入。Sentry 等のエラー監視サービスは MVP 後に検討する

**ログレベル設計**:

| レベル | 用途 |
|--------|------|
| `console.error` | システムエラー（DB 障害、予期しない例外） |
| `console.warn` | 業務エラー（重複 URL 保存、TitleFetcher 失敗） |
| `console.log` | 開発時デバッグのみ（本番コードに残さない） |

**エラーログの必須フィールド**:
```typescript
console.error(`[${appError.kind}] ${appError.message}`, { context: 'ArticleService.save', url });
```

**認証失敗ログ**（Auth.js `events.signInFailure` コールバックで出力）:
```typescript
events: {
  signInFailure({ error }) {
    console.error('[auth] signIn failed', { error: error.message });
  }
}
```

## 3. プロジェクト構成

### 3.1 コンポーネント設計パターン

**P1 導出根拠**: architecture.md §5.1 のコンポーネント一覧、設計決定 G3-1

- Atomic Design 等のフォーマルパターンは採用しない
- `src/components/` にフラット配置する
- **Server Component をデフォルト**とし、Client Component は最小限に切り出す:
  - `ArticleCard`（表示専用）: Server Component
  - `ArticleCardActions`（操作ボタン）: Client Component として切り出し、Server Actions を呼び出す
- コンポーネント間の操作はコールバック props（`onMarkAsRead`, `onMarkAsUnread`, `onDelete`）経由で受け渡す

### 3.2 ディレクトリ構成（ツリー + 役割テーブル）

**P1 導出根拠**: architecture.md §5.4 確定、設計決定 G3-2

```
src/
  app/                          # Handler 層（Next.js App Router）
    (auth)/                     # 認証関連ページ
    (app)/                      # 認証済みアプリページ
      page.tsx                  # ArticleListPage
      search/
        page.tsx                # SearchPage
    api/
      share/
        route.ts                # WebShareTargetHandler
      ping/
        route.ts                # warm-up ping エンドポイント（認証対象外）
    layout.tsx
  lib/
    services/                   # Service 層
      article-service.ts        # ArticleService（重複チェック内包）
      title-fetcher.ts          # TitleFetcher
    repositories/               # Repository 層
      article-repository.ts     # ArticleRepository
    interfaces/                 # DI 用インターフェース定義（横断的配置）
      title-fetcher.interface.ts
      article-repository.interface.ts
    db/
      schema.ts                 # Drizzle ORM スキーマ定義
      migrations/               # マイグレーション SQL ファイル
    types.ts                    # 共通型定義（ActionResult 等）
    errors.ts                   # エラー型定義（AppError 等）
  components/                   # Handler 層の再利用可能 UI コンポーネント
    article-card.tsx            # ArticleCard（表示専用）
    article-card-actions.tsx    # ArticleCardActions（操作ボタン、Client Component）
    article-save-form.tsx       # ArticleSaveForm
  middleware.ts                 # AuthGuard（Auth.js Middleware）
```

| ディレクトリ | 役割 | 対応コンポーネント |
|------------|------|------------------|
| `src/app/(auth)/` | 認証関連ページ（ログイン画面等） | — |
| `src/app/(app)/` | 認証済みアプリページ | ArticleListPage, SearchPage |
| `src/app/api/share/` | Web Share Target API ルート | WebShareTargetHandler |
| `src/app/api/ping/` | warm-up ping エンドポイント | — |
| `src/lib/services/` | Service 層（ビジネスロジック） | ArticleService, TitleFetcher |
| `src/lib/repositories/` | Repository 層（DB アクセス） | ArticleRepository |
| `src/lib/interfaces/` | DI 用インターフェース定義 | ITitleFetcher, IArticleRepository |
| `src/lib/db/` | Drizzle ORM スキーマ・マイグレーション | — |
| `src/components/` | 再利用可能 UI コンポーネント | ArticleCard, ArticleCardActions, ArticleSaveForm |
| `src/middleware.ts` | Auth.js Middleware（全ルート認証） | AuthGuard |

### 3.3 共有コード管理

**P1 導出根拠**: 設計決定 G3-3（P1 モノリシック構成）

- モノレポは採用しない。単一リポジトリ・単一アプリケーション構成とする
- 共有コードは `src/lib/` 配下に集約する:
  - 共通型定義: `src/lib/types.ts`
  - エラー型定義: `src/lib/errors.ts`
  - DI インターフェース: `src/lib/interfaces/`
  - DB スキーマ: `src/lib/db/schema.ts`

### 3.4 モジュール境界

**P1 導出根拠**: architecture.md §5.2 の依存関係設計、設計決定 G3-4

- Barrel exports（`index.ts`）は使用しない。各ファイルを直接インポートする
- **依存の方向は一方向に維持する**:
  ```
  src/app/ → src/lib/services/ → src/lib/repositories/
  ```
- **禁止する依存方向**:
  - `src/components/` から `src/lib/services/` への直接呼び出しは禁止（上位ページコンポーネントが仲介する）
  - `src/lib/repositories/` から `src/lib/services/` への依存禁止

### 3.5 設定ファイル管理（.env 戦略）

**P1 導出根拠**: NFR-4（ソースコードに認証情報を含めない）、設計決定 G3-5

| ファイル | 管理方法 | 用途 |
|---------|---------|------|
| `.env.example` | Git 管理 | 環境変数キー一覧（値は空またはダミー） |
| `.env.local` | Git 除外（`.gitignore`） | ローカル開発用の実際の値 |
| Netlify 環境変数 | Netlify コンソールで管理 | 本番環境のシークレット |

**確定済み環境変数一覧**（P1 確定、architecture.md §5.4）:

| 変数名 | 用途 |
|--------|------|
| `DATABASE_URL` | Neon PostgreSQL 接続文字列 |
| `ALLOWED_EMAIL` | シングルユーザー許可メールアドレス |
| `GOOGLE_CLIENT_ID` | Google OAuth クライアント ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth クライアントシークレット |
| `AUTH_SECRET` | Auth.js セッション署名用シークレット |
| `PING_SECRET` | warm-up ping エンドポイントの認証トークン |

### 3.6 エイリアスパス

**P1 導出根拠**: 設計決定 G3-6（Next.js デフォルト準拠）

`tsconfig.json` に以下を設定する:
```json
{
  "compilerOptions": {
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

- `@/` エイリアスのみ使用する（`@/*` → `./src/*`）
- 追加エイリアスは設定しない（小規模プロジェクトで複数エイリアスは不要）

## 4. セキュリティ規約

### 4.1 パスワードポリシー

該当なし。Google OAuth を採用しているためパスワードなし認証。独自パスワードは管理しない（P1 §7.1、設計決定 G4-1）。

### 4.2 CSRF 対策

**P1 導出根拠**: architecture.md §7.2 確定、設計決定 G4-2

| エンドポイント種別 | 対策方式 | 実装 |
|----------------|---------|------|
| Server Actions | Next.js App Router のビルトイン CSRF 保護（Origin ヘッダー検証） | フレームワーク自動提供。追加実装不要 |
| Route Handlers（WebShareTargetHandler） | Origin ヘッダー検証。不一致時 403 を返す | 手動実装。Handler 冒頭で `request.headers.get('origin')` を Netlify デプロイドメインと照合する |

**WebShareTargetHandler の CSRF 実装パターン**:
```typescript
export async function POST(request: Request) {
  const origin = request.headers.get('origin');
  // Netlify が自動提供する URL 環境変数（本番デプロイドメイン）を使用
  const allowedOrigin = process.env.URL;
  if (origin !== allowedOrigin) {
    return new Response('Forbidden', { status: 403 });
  }
  // ...
}
```

### 4.3 XSS 対策（CSP 設定含む）

**P1 導出根拠**: 設計決定 G4-3

- React / Next.js のデフォルトエスケープに依存する（JSX は自動エスケープ）
- `dangerouslySetInnerHTML` の使用を禁止する（Biome `noDangerouslySetInnerHtml` で自動強制）
- 外部リンク（記事 URL 等）には `rel="noopener noreferrer"` を付与する:
  ```tsx
  <a href={article.url} target="_blank" rel="noopener noreferrer">{article.title}</a>
  ```
- CSP ヘッダーは P3 で設定する（現時点では未設定）

### 4.4 CORS 設定

**P1 導出根拠**: 設計決定 G4-4

CORS ヘッダー設定なし。Same-Origin のみ許可する。外部ドメインからの API アクセスは想定しない。Next.js / Netlify のデフォルト動作に依存する。

### 4.5 HTTP セキュリティヘッダー

**P1 導出根拠**: architecture.md §9.6 P3 引き継ぎ、設計決定 G4-5

`next.config.js` の `headers()` で以下を設定する:

| ヘッダー | 値 | 設定方法 |
|---------|-----|---------|
| `X-Frame-Options` | `DENY` | `next.config.js` |
| `X-Content-Type-Options` | `nosniff` | `next.config.js` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | `next.config.js` |
| `Content-Security-Policy` | P3 で設定 | P3 対応 |
| `Strict-Transport-Security` | Netlify が自動設定 | プラットフォーム提供 |

### 4.6 レート制限方針

**P1 導出根拠**: 設計決定 G4-6

MVP では実装しない。認証必須（Auth.js Middleware）+ シングルユーザー構成であり、外部からの無断アクセスを構造的に防止しているため、レート制限は不要と判断する。warm-up ping エンドポイントは `PING_SECRET` トークン検証で保護する。

## 5. データ管理規約

### 5.1 スキーマ設計（正規化レベル、命名規約）

**P1 導出根拠**: architecture.md §6.2 確定、設計決定 G4-7

- **正規化レベル**: MVP はシングルテーブル構成（`articles` テーブルのみ）
- **主キー**: UUID（PostgreSQL の `gen_random_uuid()` で DB 側生成。アプリ側での生成は行わない）
- **ENUM 型**: `pgEnum` で定義する

```typescript
// src/lib/db/schema.ts
import { pgEnum, pgTable, text, timestamp, uuid } from 'drizzle-orm/pg-core';

export const statusEnum = pgEnum('status', ['unread', 'read']);

export const articles = pgTable('articles', {
  id: uuid('id').defaultRandom().primaryKey(),
  url: text('url').notNull().unique(),
  title: text('title').notNull(),
  status: statusEnum('status').notNull().default('unread'),
  saved_at: timestamp('saved_at', { withTimezone: true }).notNull().defaultNow(),
  read_at: timestamp('read_at', { withTimezone: true }),
});
```

**命名規約**:
- テーブル名: 複数形 snake_case（`articles`）
- カラム名: snake_case（`saved_at`, `read_at`）

**インデックス設計**:
- `url` カラム: UNIQUE 制約による自動 B-Tree インデックス（重複チェック用の個別インデックスは不要）
- `(status, saved_at)` 複合インデックス: 一覧表示のソート用に設定する
- タイトル・URL のキーワード検索: MVP では LIKE 部分一致（シーケンシャルスキャン）。5,000 件超または応答 1 秒超の場合は `pg_trgm` 拡張（GIN インデックス）の導入を判断する（P2 で計測・確認）

**マイグレーション管理**: Drizzle Kit によりスキーマ定義（TypeScript ファイル）からマイグレーション SQL を生成・適用する。

### 5.2 バリデーション戦略（実施レイヤー × 方式のマトリクス）

**P1 導出根拠**: architecture.md §9.3 確定、設計決定 G4-8

| レイヤー | 方式 | 目的 |
|---------|------|------|
| Handler 層（FE/BE） | Zod スキーマによる入力形式検証 | URL 形式チェック（SR-014）、空文字チェック（SR-016） |
| Service 層（BE） | ビジネスルール検証 + SSRF フィルタリング | 重複 URL チェック（SR-010）、TitleFetcher 内部 SSRF 対策 |
| DB 層 | UNIQUE 制約・NOT NULL 制約・ENUM 制約 | 並行リクエスト時のレースコンディション防止（最終保証） |

**Zod スキーマ例**:
```typescript
// 変数名: camelCase + Schema サフィックス
const articleUrlSchema = z.object({
  url: z.string().url('URLの形式が正しくありません').min(1, 'URLを入力してください'),
});
```

**原則**: バリデーション済みデータのみ Service 層へ渡す。SSRF フィルタリングは TitleFetcher 内部で実施し、Handler 層での追加フィルタリングは不要。

### 5.3 データライフサイクル

**P1 導出根拠**: SR-008（記事削除）、設計決定 G4-9

- **削除方式**: 物理削除（DELETE）を採用する。論理削除は不使用
- **保持期間**: 無期限保持
- **誤削除の対応**: Neon のポイントインタイムリストア + 定期 `pg_dump` バックアップで対応する（architecture.md §9.4）

### 5.4 トランザクション管理

**P1 導出根拠**: 設計決定 G4-10（シンプル CRUD・シングルユーザー）

- 明示的トランザクション・ロックは使用しない
- 並行リクエスト時の重複保存は DB の UNIQUE 制約で最終保証する（architecture.md §9.1）
- PostgreSQL エラーコード `23505`（unique_violation）を Repository 層で捕捉し `AppError`（`kind: 'duplicate'`）に変換する

## 6. フロントエンド UX 規約

### 6.1 アイコン・アセット管理

**P1 導出根拠**: 設計決定 G5-1（shadcn/ui エコシステム統一）

- アイコン: Lucide React（shadcn/ui デフォルト）を使用する
- PWA アイコン: `public/` ディレクトリに配置する
- 画像等の静的アセット: `public/` に配置し、`next/image` コンポーネント経由で利用する

### 6.2 レイアウト・ページテンプレート

**P1 導出根拠**: architecture.md §5.4 確定、設計決定 G5-2

3 種類のレイアウトを使用する（Next.js App Router 規約準拠）:

| レイアウト | 適用範囲 | ファイルパス |
|----------|---------|-----------|
| ルートレイアウト | 全ページ共通 | `src/app/layout.tsx` |
| `(auth)` グループ | 認証関連ページ | `src/app/(auth)/layout.tsx` |
| `(app)` グループ | 認証済みアプリページ | `src/app/(app)/layout.tsx` |

### 6.3 アクセシビリティ（WCAG 準拠レベル）

**P1 導出根拠**: 設計決定 G5-3

- WCAG 2.1 Level A を意識する（個人ツールのため必須達成レベルとしては設定しない）
- shadcn/ui（Radix UI ベース）のデフォルトアクセシビリティ対応に依存する
- 追加対応コストが小さい範囲で実施する（WAI-ARIA ロールは shadcn/ui が付与）

### 6.4 メタデータ・SEO

**P1 導出根拠**: 設計決定 G5-4（認証必須の個人ツール）

- SEO 対応は不要（認証必須の個人ツールのため）
- 最低限の `<title>` と `<meta name="description">` を設定する
- `public/robots.txt` で全クローラーを拒否する:
  ```
  User-agent: *
  Disallow: /
  ```
- PWA manifest（`public/manifest.json`）を配置する

### 6.5 レスポンシブ設計（ブレイクポイント定義含む）

**P1 導出根拠**: IR-002（スマートフォン・PC 両対応）、設計決定 G5-5

- モバイルファーストで実装する（スマートフォン表示を基準に PC 向けを拡張）
- Tailwind CSS のデフォルトブレイクポイントを使用する（追加定義なし）:

| ブレイクポイント | 幅 | 用途 |
|--------------|-----|------|
| デフォルト（`sm:` 未満） | < 640px | スマートフォン |
| `sm:` | ≥ 640px | — |
| `md:` | ≥ 768px | タブレット（参考） |
| `lg:` | ≥ 1024px | PC |

- スマートフォン + PC の 2 パターンを主に考慮する

### 6.6 タッチ・マウス両立

**P1 導出根拠**: 設計決定 G5-6（Radix UI 標準対応）

- 最小タッチターゲットサイズ: 44×44px（iOS HIG / WCAG 準拠）
- shadcn/ui（Radix UI）のデフォルトイベント処理（`onClick`）に依存する
- タッチとマウスの個別対応は不要（Radix UI が両対応済み）

### 6.7 ビューポート・Safe Area

**P1 導出根拠**: 設計決定 G5-7（Android PWA 必須対応）

`app/layout.tsx` の `<head>` に以下を設定する:
```html
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
```

CSS で Safe Area Inset を考慮する:
```css
padding-bottom: env(safe-area-inset-bottom);
padding-top: env(safe-area-inset-top);
```

## 7. AI 統合規約

省略（AI 統合は MVP 後に決定）。

AI 機能（UC-11〜UC-16）は MVP スコープ外（requirements.md §5.3）。プロンプト管理・レート制限・リトライ戦略の規約は AI 機能採用時に設計する。

## 8. 規約の自動強制サマリ

| 規約 | 強制手段 | タイミング |
|------|---------|-----------|
| Lint（コード品質ルール） | Biome | 保存時（エディタ）/ commit 時（lefthook）/ CI |
| フォーマット（コードスタイル） | Biome | 保存時（エディタ）/ commit 時（lefthook）/ CI |
| インポート順序 | Biome `organizeImports` | 保存時（自動）|
| `var` 禁止 | Biome `noVar` | 保存時（自動）|
| `const` 優先 | Biome `useConst` | 保存時（自動）|
| `dangerouslySetInnerHTML` 禁止 | Biome `noDangerouslySetInnerHtml` | 保存時（自動）|
| TypeScript `strict` / `noUncheckedIndexedAccess` | TypeScript コンパイラ | ビルド時 / 保存時（エディタ）|
| パスエイリアス `@/` | `tsconfig.json` | コンパイル時（自動）|
| `.env.local` の Git 除外 | `.gitignore` | commit 時（自動）|
| CSRF 保護（Server Actions） | Next.js App Router 組み込み | リクエスト時（自動）|
| CORS 拒否（デフォルト） | Next.js / Netlify デフォルト | リクエスト時（自動）|
| HTTP セキュリティヘッダー（X-Frame-Options 等） | `next.config.js` | リクエスト時（自動）|
| DB UNIQUE 制約（URL 重複防止） | Neon PostgreSQL / Drizzle スキーマ | DB 書き込み時（自動）|
| DB NOT NULL / ENUM 制約 | Neon PostgreSQL / Drizzle スキーマ | DB 書き込み時（自動）|
| 命名規則（一部：Biome カバー範囲） | Biome 命名ルール（部分的） | 保存時（自動、一部）|
| 命名規則（ファイル名・DB カラム名等） | 手動遵守 | コードレビュー時 |
| エラーハンドリングパターン（`AppError`/`ActionResult`） | 手動遵守（型システムが部分サポート） | コードレビュー時 |
| 非同期処理（`async/await` 統一） | 手動遵守 | コードレビュー時 |
| コメント方針（Why のみ） | 手動遵守 | コードレビュー時 |
| Git commit メッセージ（Conventional Commits） | 手動遵守 | commit 時 |
| モジュール境界（依存方向の一方向維持） | 手動遵守 | コードレビュー時 |
| ログ出力方針（レベル設計） | 手動遵守 | コードレビュー時 |
| タッチターゲット 44px | 手動遵守 | コードレビュー時 |

**自動強制可能な規約**: 14 件
**手動遵守のみの規約**: 10 件
**合計**: 24 件

## 9. P3 への引き継ぎ事項

P3（開発プロセス設計）に引き継ぐ横断的関心事（architecture.md §9.6 より）:

| トピック | 引き継ぎ内容 |
|---------|------------|
| セキュリティ自動化 | ソースコード内認証情報スキャンの自動化（CI での `pnpm audit` 等）、依存関係脆弱性スキャン運用 |
| CSP 設定 | `Content-Security-Policy` ヘッダーの設定（現時点では未設定、P3 で設計・実装） |
| パフォーマンス計測 | CI/CD への Lighthouse スコア・Core Web Vitals 計測の組み込み |
| マイグレーション運用 | マイグレーション失敗時のロールバック手順の詳細化、定期リストアテストの手順整備 |
| バックアップ運用 | `pg_dump` バックアップの定期実行手順（週次）。**初回本番デプロイ前に整備すること** |
| Drizzle バージョン管理 | major バージョンアップデート対応の運用ルール、changelog 購読手順 |
| warm-up ping 運用 | Netlify Starter プランでの Scheduled Functions 利用可否確認。利用不可の場合は代替手段（GitHub Actions scheduled workflow 等）の選択と実装 |

## 10. P1 トレーサビリティ

| P1 決定 | 導出された P2 規約 |
|--------|-------------------|
| ADR-0001: Netlify（Starter）採用 | §3.5 環境変数管理（Netlify コンソール管理）、§4.6 レート制限なし方針、§9 P3 引き継ぎ（warm-up ping 運用）|
| ADR-0002: Next.js + Node.js + pnpm 採用 | §1.2 TypeScript strict 設定、§2.2 async/await 統一、§3.1 Server Component デフォルト、§3.2 ディレクトリ構成（app/ Handler 層）、§3.6 `@/` エイリアス |
| ADR-0003: Neon (PostgreSQL) + Drizzle ORM 採用 | §5.1 スキーマ設計（pgEnum・UUID・UNIQUE 制約）、§5.2 バリデーション戦略（DB 層の UNIQUE/NOT NULL）、§5.4 トランザクション管理（UNIQUE 最終保証）|
| ADR-0004: Biome 単独採用 | §1.1 Linter/Formatter 設定、§1.5 インポート順序（`organizeImports`）、§2.3 `var` 禁止・`const` 優先、§4.3 `dangerouslySetInnerHTML` 禁止（Biome ルール）|
| architecture.md §5.4: ファイル配置・命名規則確定 | §1.3 命名規則全般、§3.2 ディレクトリ構成詳細、§3.3 共有コード管理 |
| architecture.md §7.2: 認証・アクセス制御方針確定 | §4.1 パスワードポリシー N/A、§4.2 CSRF 対策（Server Actions ビルトイン + Route Handler Origin 検証）、§4.4 CORS なし |
| architecture.md §9.1: エラーハンドリング方針確定 | §2.1 エラーハンドリング（AppError・ActionResult パターン・各層の役割）|
| architecture.md §9.3: バリデーション方針確定（Zod） | §5.2 バリデーション戦略（Zod スキーマ + DB 制約の二段構え）|
| architecture.md §9.4/9.5: バックアップ・Neon コールドスタート対策 | §9 P3 引き継ぎ（pg_dump 週次・warm-up ping）|
| requirements.md §5.3: AI 機能 MVP スコープ外 | §7 AI 統合規約（省略）|
| requirements.md NFR-4: セキュリティ要件 | §4.3 XSS 対策、§4.5 HTTP セキュリティヘッダー、§3.5 .env 管理 |
| requirements.md IR-002: スマートフォン・PC 両対応 | §6.5 レスポンシブ設計（モバイルファースト）、§6.6 タッチ・マウス両立 |
