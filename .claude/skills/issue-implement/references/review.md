# レビュー詳細手順

レビューは2段階で実施する: Phase 2（設計レビュー）と Phase 5（最終レビュー）。

---

## 設計レビュー（Phase 2）

I/F定義のみを対象にレビューする。Codex（汎用）と Opus（プロジェクト固有）を **並列実行** し、別 LLM の視点で Opus の自己正当化バイアスを補完する。

### 2a. Codex 汎用レビュー（Codex CLI）

`git diff main` を取得し、以下のプロンプトを `printf "%s" "$prompt" | timeout 120 codex exec --full-auto -` で渡す:

```
以下は I/F 設計（型・インターフェース・関数シグネチャ・動作コメント）の diff です。
実装本体はまだありません（panic("not implemented") や TODO placeholder）。
この I/F 設計をレビューしてください。観点は自由です。

レビュー結果は以下の JSON 形式で .output/issue-implement/{issue_number}/review-design-codex.json に書き出してください:
{
  "findings": [
    {
      "file": "ファイルパス",
      "line": 行番号,
      "severity": "error|warning|info",
      "category": "カテゴリ（null-safety, dry, naming, best-practice, type-safety 等）",
      "message": "指摘内容",
      "suggestion": "修正案"
    }
  ]
}

{git diff main の内容}
```

観点を縛らないことで、Opus が見落としがちな問題を拾う:
- nullable フィールドの扱い漏れ（pgtype.Timestamptz の Valid チェック等）
- DRY 違反になりそうな構造
- フレームワーク固有のベストプラクティス
- コンパイル時安全性（型アサーション `var _ Interface = (*Impl)(nil)` 等）

### 2b. Opus プロジェクト固有レビュー（サブエージェント）

プロジェクト固有の知識が必要な観点を Opus サブエージェントで確認する。

**サブエージェントに渡す情報:**
- Phase 1 で定義した型・インターフェース・シグネチャ・動作コメント（git diff で取得）
- 作業ディレクトリパス（既存コード読み取り用）
- 出力先: `.output/issue-implement/{issue_number}/review-design-opus.json`
- 「Write で書き出すこと」
- Codex と同じ JSON フォーマットで出力すること（findings 配列 with file, line, severity, category, message, suggestion）

**確認すべき観点:**
- **既存パターン適合**: 周辺の既存コードと命名・構成が一致しているか（既存コードを Read して比較すること）
- **ADR 準拠**: `docs/adr/` の決定事項に反していないか

### 指摘の統合と修正（投票ルール）

両レビューの JSON 出力を Read で読み込み、以下のルールで統合する:
- **Codex と Opus の両方が同一箇所（同一ファイル・同一カテゴリ）を指摘** → 自動採用
- **片方のみの指摘** → メインエージェントが妥当性を判断

メインエージェントが採用された指摘で I/F定義を修正し、Phase 3（テスト作成）に進む。

---

## 最終レビュー（Phase 5）

実装完了後、多観点でコードレビューを実施する。Phase 2 の設計レビューで I/F レベルの問題は解消済みのため、実装レベルの問題に集中する。

### 5a. LLM 特有の前処理 + 静的解析

レビュー前に機械的に検出・修正する:
- **API 幻覚の検出**: 変更対象の Go サービスで `go mod tidy` を実行し、go.mod/go.sum に差分がないか確認
- **静的解析**: 変更対象の Go サービスで `go vet ./...` を実行。AI 生成コードは並行性エラーが人間の 2 倍との報告がある（CodeRabbit 2025 分析、vendor 調査）ため推奨
- **不完全な実装の検出**: `git diff main` の変更ファイルに TODO/FIXME/HACK/XXX が残存していないか確認

### 5b. レビュー実行

以下を並列で実行する。各サブエージェントには `git diff main` の差分と作業ディレクトリパスを渡し、結果を `.output/issue-implement/{issue_number}/` 配下に **JSON 形式** で書き出すよう指示する。Phase 2 と同じ `findings` 配列フォーマット（file, line, severity, category, message, suggestion）を使用し、採用判定の入力を安定化する。

#### 1. Codex Review（別LLM）

`/codex-review --diff main` を Skill ツールで実行する。codex CLI が利用できない場合はスキップ。

Claude が書いたコードを Claude 自身がレビューすると自分の判断を正当化するバイアスが働くため、別 LLM の目を入れて補完する。

#### 2. ゴール整合性レビュー（サブエージェント）

Issue のコンテキスト（受け入れ条件・スコープ外・計画）を持つエージェントでないと判断できない観点。

**追加で渡す情報:**
- Issue 本文（受け入れ条件・スコープ外）
- 計画の実装ステップ

**確認すべき観点:**
- 受け入れ条件の各項目が実装でカバーされているか
- Issue の「スコープ外」に記載された内容に触れていないか
- 計画にない追加機能・リファクタリングが含まれていないか
- 受け入れ条件の独自解釈がないか
- `.output/issue-implement/{issue_number}/if-changes.json` が存在する場合、Phase 4 で発生した I/F 変更が受け入れ条件との整合性を損なっていないか確認する

#### 3. 変更容易性レビュー — 実装レベル（サブエージェント）

Phase 2 で I/F レベルは確認済み。ここでは実装レベルの問題を確認する。

**確認すべき観点:**

LLM アンチパターン（`docs/guides/architecture/` の各ガイドを参照）:
- **過剰抽象化** (`over-abstraction.md`): 1実装しかないのにインターフェースを切っている（Phase 1 で定義したものは除く）、1箇所でしか使わないヘルパー関数を作っている
- **設定の過剰外部化** (`over-externalization.md`): 固定で良い値を環境変数や設定ファイルに切り出している
- **形式的エラーハンドリング** (`cargo-cult-error-handling.md`): 意味のない err wrap、catch して re-throw するだけ、冗長な nil チェック

実装品質:
- 既存コードのエラーハンドリング・ログ出力パターンと一致しているか
- 不要な変数・関数が残っていないか

#### 4. セキュリティ深掘り（条件付き）

**発動条件:** 認証・認可、トークン・セッション処理、ユーザー入力処理、API エンドポイント追加、データアクセス制御のいずれかを含む場合。

このプロジェクトは Ory Hydra + JWT を使った認証基盤がある（ADR 参照）。プロジェクト固有の認証フローとの整合性を確認すること。

#### 5. パフォーマンス深掘り（条件付き）

**発動条件:** DB クエリ追加・変更、大量データ処理、ループ内外部呼び出しのいずれかを含む場合。

sqlc 生成コードの使い方、Connect-RPC の特性を踏まえた判断が求められる。

### 5c. 採用判定（別サブエージェント）

全レビュー結果を集約し、`agents/review-filter.md` を Read で読み込んで採用判定サブエージェントを起動する。

レビュアとは独立したコンテキストで実行する。各指摘をコードの事実に基づいて独立に判断する。

### 5d. 修正 & テスト再実行

採用された指摘に対応してコードを修正し、テストが壊れていないか確認する。
