---
name: design
description: 対話的に画面デザインを作成・修正し、完了後に Issue 化して /issue-plan → /issue-implement で実装するワークフロー。EC サイト（React + Tailwind）と管理画面（templ + HTMX）の両方に対応。`/design` で起動。「デザイン作って」「画面作りたい」「UIを変えたい」等と言った場合にも使用する。
allowed-tools: Bash, Read, Glob, Grep, Agent, Write, Edit, AskUserQuestion
user-invocable: true
argument-description: "対象画面の説明。例: '商品一覧ページ', 'カート画面のリデザイン', 'トップページ'"
---

# design

対話的に画面デザインを作成・修正し、完了後に GitHub Issue を起票する。デザイン→実装の流れは `design → Issue化 → issue-plan → issue-implement` で一貫させる。

**重要: ユーザーとの対話および Issue への書き込みは日本語で行うこと。**

### ユーザー対話のルール

- **AskUserQuestion は「固定選択肢から1つ選ぶだけ」の場面でのみ使用する**（例: EC サイト / 管理画面 のどちらですか？）
- それ以外の対話（自由入力、承認+フィードバック、確認）は **テキスト出力のみ** で行い、ユーザーの入力を待つ
- 「問題なければOKと入力してください。修正があればお願いします。」のように、承認とフィードバックの両方を受け付ける形が望ましい

## ワークフロー

```
Phase 0: 初期化     → 対象画面の特定、前提確認、コンテキスト読み込み
Phase 1: デザイン生成 → anti-AI ルール適用、初版デザインの作成
Phase 2: 対話的修正  → ユーザーフィードバックに基づく修正ループ
Phase 3: Issue 起票  → 成果物書き出し、Issue 作成、ファイル復元
```

---

## Phase 0: 初期化

### 0a. 前提チェック

1. `gh auth status` で GitHub CLI の利用可能性を確認する。失敗時は案内して終了

2. Tailwind CSS の導入状況を確認する:
   - EC サイト: `services/ec-site/frontend/src/index.css` に `@import "tailwindcss"` が存在するか
   - 管理画面: Tailwind の設定ファイルが存在するか

   未導入の場合は以下を表示して終了:
   ```
   対象の画面に Tailwind CSS が導入されていません。
   先に Tailwind CSS を導入してください。
   ```

### 0b. 引数の解析

| パターン | 動作 |
|---------|------|
| テキストあり | 対象画面の説明として Phase 0c に進む |
| 引数なし | `どの画面のデザインを作成しますか？` と表示して入力を待つ |

### 0c. 対象画面の特定

ユーザーの入力から対象画面を特定し、出力形式を決定する。

| 対象 | 出力形式 | 根拠 |
|------|---------|------|
| EC サイト（商品一覧、カート、注文等） | React (.tsx) + Tailwind CSS v4 | ADR-0001, ADR-0017 |
| 管理画面（商品管理、顧客管理） | templ (.templ) + Tailwind CSS | ADR-0007 |

判断が曖昧な場合は AskUserQuestion で確認する。

### 0d. コンテキストの読み込み

対象画面に応じて以下を読み込む:

**EC サイトの場合:**
- `services/ec-site/frontend/src/index.css` — `@theme` セクションからデザイントークンを抽出
- `services/ec-site/frontend/src/` 配下の既存コンポーネント — レイアウト・スタイルの一貫性を維持するため

**管理画面の場合:**
- `services/{service-name}/templates/layout.templ` — 共通レイアウトの構造
- `services/{service-name}/templates/` 配下の既存 `.templ` ファイル — 参照コンテキストとして渡し、templ 生成精度を向上させる
- Tailwind 設定ファイル（存在する場合）

**共通:**
- 対象画面に関連する proto ファイル（`proto/` 配下）— Phase 2 の機能要件チェックで使用
- 対象画面に関連する sqlc クエリ（`db/query/` 配下）— Phase 2 の機能要件チェックで使用

### 0e. デザイントークンの確認

読み込んだデザイントークンをユーザーに提示する。

**EC サイトの場合（`@theme` から抽出）:**
```
現在のデザイントークン:
- カラー: background=#ffffff, foreground=#111827, surface=#e5e7eb, muted=#6b7280, border=#d1d5db
- フォント: "Inter", ui-sans-serif, system-ui, sans-serif

これらのトークンを基にデザインします。変更したい場合はお知らせください。
```

トークンが最小限しか定義されていない場合は、デザインに必要な追加トークン（primary カラー、accent カラー等）の定義をユーザーに提案する。

---

## Phase 1: デザイン生成

### 1a. dev server の起動確認

`just dev` による live reload が利用可能であることを確認する。起動していない場合はユーザーに案内する:

```
デザインの確認のため、dev server を起動してください:
just dev

起動したらOKと入力してください。
```

### 1b. ビジュアルスタイルの選択

`references/anti-ai-design.md` を Read で読み込み、ルールを適用する。

ユーザーがビジュアルスタイルを指定していない場合は、対象画面に適した 2〜3 の候補を提示して選択してもらう:

```
デザインのビジュアルスタイルを選んでください:

1. エディトリアル — 雑誌的レイアウト、強い見出し、情報の階層化が明確
2. スイスデザイン — グリッドの厳格さ、サンセリフ、ミニマルだが機能的
3. その他（具体的に指定）

番号で選んでください。
```

### 1c. デザインの段階的生成

anti-AI ルールの「段階的詳細化」に従い、以下の順序で生成する:

1. **構造**: ページ構成・セクション配置を作成し、ファイルに書き出す
2. ユーザーにブラウザで確認してもらい、承認を得る
3. **配色・タイポグラフィ・スペーシング**: デザイントークンに基づいてスタイルを適用する
4. ユーザーに確認してもらい、承認を得る

**templ コンポーネント生成時の制約（管理画面の場合）:**
- sqlc 生成型（`db.Product` 等）を templ コンポーネントに直接渡さない。プレゼンテーション用の型に変換してから渡す（`.claude/rules/db-type-boundary.md` 準拠）
- 既存の `.templ` ファイルの構文・パターンに合わせる

### 1d. 生成したファイル

feature branch 上で直接ファイルを編集する。新規ファイルの作成と既存ファイルの編集の両方が発生しうる。

---

## Phase 2: 対話的修正

### 2a. フィードバックループ

ユーザーが「完了」と言うまで以下を繰り返す:

1. ユーザーからフィードバックを受け取る
2. anti-AI ルールに準拠しているか確認しつつ、差分編集を適用する
3. 変更内容を説明し、ブラウザで確認してもらう

修正時のルール（`references/anti-ai-design.md` セクション4）:
- 1回の修正で1つの変更に絞る
- 変更しない部分を明示する
- 変更前後で何が変わったかを説明する

### 2b. 機能要件チェック

ユーザーが「完了」と言った時点で、proto/DB スキーマとの整合性をチェックする。proto/DB スキーマが対象画面に存在しない場合はスキップする。

**チェック手順:**

1. 対象画面に関連する proto ファイルから message フィールドと rpc 一覧を抽出する:

```bash
# proto から message フィールドを抽出
grep -E '^\s+(string|int|bool|repeated|google)' proto/{service}/v1/*.proto

# rpc 一覧を抽出
grep -E '^\s+rpc ' proto/{service}/v1/*.proto
```

2. 対象画面に関連する sqlc クエリから SELECT カラムを抽出する（管理画面の場合）:

```bash
# sqlc クエリのカラムを確認
grep -A5 'SELECT' services/{service}/db/query/*.sql
```

3. デザインが表示すべきフィールドを漏れなくカバーしているかチェックリストを作成する:

```
機能要件チェック:
- [x] product.name — 商品名が表示されている
- [x] product.price — 価格が表示されている
- [ ] product.created_at — 作成日時が表示されていない

不足しているフィールドがあります。追加しますか？
```

4. 不足があればユーザーに確認し、必要に応じて修正する

---

## Phase 3: Issue 起票

`references/issue-creation.md` に従い、Issue の起票とファイルの復元を行う。

### 手順

1. `.output/design/{session_id}/` にデザイン判断の記録を書き出す
2. Issue 本文を作成してユーザーに提示する
3. 承認後、`gh issue create` で Issue を起票する
4. 作業ブランチのデザイン作業ファイルを復元する（対象サービスのディレクトリに限定。`references/issue-creation.md` セクション3参照）
5. ユーザーに Issue 番号と次のステップを案内する

**重要:** ファイル復元（`git checkout -- .`）の実行前に必ずユーザーの承認を得ること。
