### タスク001: ツールチェーンセットアップ

**blocked_by**: なし

#### 成果物

- [x] Biome 設定（biome.json: recommended ルールセット、noVar・useConst・noDangerouslySetInnerHtml 有効化、organizeImports 有効化）
- [x] TypeScript 設定（tsconfig.json: strict: true、noUncheckedIndexedAccess: true、@/ パスエイリアス設定）
- [x] VS Code 設定ファイル一式（.vscode/settings.json: Biome 拡張有効化・保存時自動フォーマット設定）
- [x] VS Code 拡張機能リスト（.vscode/extensions.json: Biome、Playwright Test Explorer 等）
- [x] VS Code デバッグ設定（.vscode/launch.json: Next.js サーバーサイドデバッグ用 Node.js アタッチ設定）

#### 受け入れ条件

- [x] `pnpm biome check` がエラーなしで通過する
- [x] `tsc --noEmit` が型エラーなしで通過する
- [x] .vscode/settings.json、.vscode/extensions.json、.vscode/launch.json の 3 ファイルが存在する
- [x] biome.json に `recommended: true`、`noVar`、`useConst`、`noDangerouslySetInnerHtml`、`organizeImports` が設定されている

#### 入力

- `standards.md` §1.1（Biome 設定方針）
- `standards.md` §1.2（TypeScript 設定）
- `standards.md` §3.6（エイリアスパス）
- `development-process.md` §5.1（デバッグ環境・VS Code 設定ファイル）
