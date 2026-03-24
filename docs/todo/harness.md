# Harness Engineering 導入検討メモ

対象記事: [Claude Code / Codex ユーザーのための誰でもわかるHarness Engineeringベストプラクティス](https://nyosegawa.com/posts/harness-engineering-best-practices-2026/)

## 目的

記事で提案されている事項を、このリポジトリの現状と照らして整理し、未導入項目の導入是非を判断しやすくする。

## 先に結論

このリポジトリは、Harness Engineering の土台はかなり整っている。

- `CLAUDE.md` が短いポインタとして機能している
- ADR が `docs/adr/` に整理されている
- `lefthook` が導入済み
- `Justfile` に lint / format / test / generate / dev の入口がまとまっている
- `.claude/settings.json` で `PostToolUse` / `PreToolUse` / `Stop` Hookが設定されている

一方で、記事の中でも効果が大きいとされる次の項目は、未導入または弱い。

- E2E テスト
- CI ワークフロー
- TypeScript 型チェックの常設
- Codex 向け `AGENTS.md`
- セッション間状態管理の標準化

## 現在できていること

### 1. リポジトリ衛生

- ADR が `docs/adr/` に整理されている
- リポジトリのルールや実装パターンへの参照先が `CLAUDE.md` にまとまっている
- lint / format / secret scan の入口が `Justfile` に定義されている

### 2. 決定論的ツール

- TypeScript は `oxlint` / `oxfmt`
- Go は `golangci-lint`
- secret scan は `gitleaks`
- `lefthook` で pre-commit / pre-push を実行

### 3. Hooks

- `PostToolUse`: 編集後に TS / Go の format + lint
- `PreToolUse`: `oxlint.json` と `.golangci.yml` の編集をブロック
- `Stop`: セッション完了時に `just lint` を強制

### 4. ポインタ型の `CLAUDE.md`

- 内容は短く、役割も明確
- 記事が推奨する「詳細ではなく参照先だけを書く」にかなり沿っている

## 未導入または弱い項目

| 項目 | 現状 | 効果 | 推奨 | 判断メモ |
|---|---|---:|---:|---|
| E2E テスト導入 | ほぼ未導入。確認できたテストは一部の Go テストのみ | 大 | 最優先 | 記事の「完了をテストで検証する」の中核。Web があるので Playwright CLI が第一候補 |
| CI ワークフロー | `.github/workflows` が見当たらない | 大 | 最優先 | hooks だけでは取りこぼしがある。遅い層の品質ゲートが必要 |
| TypeScript 型チェックの常設 | `build` では `tsc` を使うが、日常の `lint` / hook に未接続 | 大 | 強く推奨 | TS 系の回帰を早いフィードバックで止められる |
| Codex 向け `AGENTS.md` | `CLAUDE.md` のみで、`AGENTS.md` は未作成 | 中 | 強く推奨 | Codex 系でも同じ共通指示を読ませやすくなる |
| Stop Hook の完了条件強化 | 現在は `just lint` のみ | 中 | 推奨 | `just test` や将来の E2E を完了条件に含めたい |
| セッション間起動ルーチンの標準化 | `just dev` はあるが、開始手順は標準化されていない | 中 | 推奨 | `git log` 確認、進捗読込、疎通確認を毎回同じ流れにしたい |
| JSON ベース進捗記録 | `docs/plans/*.md` はあるが JSON 進捗ファイルは未整備 | 中 | 推奨 | セッション継続のための構造化状態がない |
| カスタム lint / architecture rule | 汎用 lint はあるが、境界違反用の独自ルールは未確認 | 大 | 推奨 | ADR を実行可能ルールに変えると効果が大きい |
| ADR と lint の結合 | ADR はあるが、機械的検証への接続は未確認 | 大 | 推奨 | `archgate` 的な発展先。再発防止に効く |
| 保護対象設定ファイルの拡張 | `oxlint.json` と `.golangci.yml` のみ保護 | 中 | 推奨 | `lefthook.yml`、CI、`tsconfig` 等も保護候補 |
| 記述的ドキュメントの整理 | `docs/plans/*.md` は詳細で、将来的に腐敗リスクがある | 中 | 推奨 | 計画資料としては有用だが、真実のソース化は避けたい |
| ガベージコレクション運用 | 未導入 | 低〜中 | 後回し | 今の規模では優先度は低い |
| ハーネス効果の定量測定 | 未導入 | 中 | 後回し | 運用が乗ってから導入で十分 |

## 導入優先度の提案

### すぐやる価値が高い

1. E2E テスト導入
2. CI ワークフロー導入
3. TypeScript 型チェックを `lint` / hook に組み込む
4. `AGENTS.md` を追加して Codex 向け共通指示を整備する
5. `Stop` Hook の完了条件を `lint` だけでなく `test` まで広げる

### 次の段階でやると効く

1. セッション開始ルーチンの標準化
2. JSON ベースの進捗ファイル導入
3. ADR と lint の結合
4. アーキテクチャ境界を検査するカスタムルール追加
5. 保護対象設定ファイルの拡張

### 後回しでよい

1. ガベージコレクションエージェント
2. ハーネス効果の定量測定
3. 高度なフィードバックループ

## 項目ごとの補足

### E2E テスト

記事の主張の中で、このリポジトリへの影響が最も大きい候補。  
現状は unit / integration もまだ薄く、UI やサービス疎通を「完了条件」として機械的に判定できない。

導入するなら候補は次の順。

1. Playwright CLI
2. 将来的に visual regression
3. 必要なら API レベルの疎通テスト

### CI ワークフロー

記事の「フィードバック速度の層構造」に照らすと、今は local hook はあるが CI 層が弱い。  
最低限でも次を GitHub Actions で毎回回したい。

- `just lint`
- `just test`
- 必要なら `just fmt-check`

### TypeScript 型チェック

現状の TypeScript パッケージでは `build` 時にしか `tsc` が見えない。  
記事の思想では、型エラーはできるだけ速い層に寄せるべきなので、`lint-ts` または専用 `typecheck` を追加し、hook / CI の両方に載せたい。

### `AGENTS.md`

このリポジトリは Claude 側の整備が進んでいる。  
一方で記事は、共通ハーネスレイヤーとして `AGENTS.md` を持つことを強く勧めている。  
内容は `CLAUDE.md` を複製するだけでなく、Codex でも必要な最小限のポインタに絞るのが良い。

### セッション間状態管理

今の `docs/plans/*.md` は計画資料としては有用だが、記事で言う「各セッションが起動時に必ず読む構造化された進捗状態」にはなっていない。  
候補としては次のような運用がある。

- `docs/todo/progress.json` を置く
- 現在のタスク、次にやること、ブロッカーだけを保持する
- セッション開始時に `git log --oneline -20` と一緒に読む

### ADR と lint の結合

今は ADR がある一方で、その決定がコードで破られた時に自動検知する仕組みは弱い。  
ここがつながると、記事のいう「決定を仕組みで強制する」に一段近づく。

## 導入判断のおすすめ

現時点では、次の順で検討するのがバランスがよい。

1. E2E
2. CI
3. TypeScript 型チェック常設
4. `AGENTS.md`
5. `Stop` Hook 強化
6. セッション状態管理
7. ADR と lint の接続

逆に、ガベージコレクションや効果測定は、今すぐ入れなくてもよい。

## 根拠として確認した現状

- `CLAUDE.md` は短いポインタ型
- `.claude/settings.json` に `PostToolUse` / `PreToolUse` / `Stop` がある
- `Justfile` に `lint` / `test` / `fmt-check` / `dev` がある
- `lefthook.yml` に pre-commit / pre-push がある
- `docs/adr/` に ADR が継続的にある
- `.github/workflows` は未確認
- E2E テスト設定ファイルは未確認
- `docs/plans/*.md` はあるが JSON 進捗管理は未確認
