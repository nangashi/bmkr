# 実装

テスト Green → wip クリーンアップ → 品質ゲート（ベースライン確認 + Codex レビュー + simplify + コメント整理）の順で進める。

---

## Step 1: 実装ループ

`references/implementation-loop.md` に従い、スタブ（`panic("not implemented")`）を実装してテストを Green にする。

---

## Step 2: `wip:` マーカーのクリーンアップ

`grep -rn "// wip:" .` で全マーカーを検出し、各行を以下のルールで処理する:

- **新規関数の godoc コメント**: `// wip:` プレフィックスを除去し、正式な godoc として残す
- **既存コード内のインラインコメント**: コードから自明な内容は削除、付加価値のある説明のみ `// wip:` を除去して残す

---

## Step 3: 品質ゲート

### ベースライン確認

`just test` + `just fmt` + `just lint` を実行する。失敗時はエラーを分析して修正（最大3回）。3回で解決しない場合は中断。

実装ループ経由の場合も lint が未解決の可能性があるため、ここで保証する。

### レビュー・修正ループ（最大3回、Round 3 は条件付き）

```
# 品質ゲート開始前のチェックポイント
git add -A && git commit -m "checkpoint: quality-gate-pre (issue #{number})"

for round in 1..3:
  # Round 3 は振動検出時のみ実行
  if round == 3 and oscillation_directives が空:
    break

  Codex 観点付きレビュー → 採用判定（directive があれば filter に渡す）
  採用指摘なし → ループ終了

  Codex 修正（fix + test + fmt + lint、directive があればプロンプトに注入）
  git add -A && git commit -m "checkpoint: quality-gate-round-{round}"

  # Round 2 以降: 振動検出
  if round >= 2:
    振動検出を実行（後述）

simplify（冗長コード整理）
コメント整理（simplify 後に1回）
```

### Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 300 codex exec --full-auto` に stdin で渡す。codex CLI が利用できない場合はスキップ。

- `{diff_command}`: `git diff main`
- `{perspective_files}`: `review-silent-failure.md`, `review-project-compliance.md`, `review-security.md`
- `{output_path}`: `.output/issue-implement/{issue_number}/review-final.md`

Claude が書いたコードを Claude 自身がレビューすると自分の判断を正当化するバイアスが働くため、別 LLM の目を入れて補完する。

### 採用判定（別サブエージェント）

`agents/review-filter.md` を Read で読み込んだ Haiku モデルの採用判定サブエージェント（`model: haiku`）を起動する。レビュアとは独立したコンテキストで、各指摘をコードの事実に基づいて判断する。分類・フィルタリングタスクのため Haiku で十分な精度が得られる。

渡すパラメータ: `issue_number`、`review_output_path`（`.output/issue-implement/{issue_number}/review-final.md`）、`output_path`（`.output/issue-implement/{issue_number}/review-final-filtered.md`）。Issue 本文や oscillation directives はサブエージェントが自己取得する。

採用指摘がなければループ終了。

### 修正（Codex）

採用された指摘と `git diff main` を Codex に渡し、修正を実行させる。Codex には以下を指示する:

- 採用された指摘に対応してコードを修正すること
- 修正後に `just test`、`just fmt`、`just lint` を実行して問題がないことを確認すること
- `.output/issue-implement/{issue_number}/oscillation-directives.md` が存在する場合、その内容を「振動回避指示」としてプロンプトに含める。directive に記載された変更は再度行わないこと

### 振動検出（Round 2 以降、オーケストレータが実行）

各ラウンドの Codex 修正後、コミットチェックポイントを作成し、振動を検出する。

#### チェックポイント

```
git add -A && git commit -m "checkpoint: quality-gate-round-{round}"
```

#### 検出手順

1. `git diff HEAD~2..HEAD~1`（前ラウンドの変更）を取得
2. `git diff HEAD~1..HEAD`（今ラウンドの変更）を取得
3. 同一ファイルの同一関数・同一ブロック内で、前ラウンドの変更を打ち消す変更がないか確認する

**振動の判定基準:**
- 前ラウンドで追加された行が今ラウンドで削除されている
- 前ラウンドで削除された行が今ラウンドで復元されている
- 前ラウンドで変更されたロジックが今ラウンドで元に戻されている

#### 振動検出時の対応

1. 両バージョン（前ラウンドの状態 / 今ラウンドの状態）のコードを比較し、採用判定で accept された指摘の意図に沿う方を判定する
2. 劣った方の変更を `git checkout HEAD~1 -- {file}` で部分的に revert する（振動していない変更は保持する）
3. directive を `.output/issue-implement/{issue_number}/oscillation-directives.md` に追記する:

```markdown
## Directive {番号}
- ファイル: {file}
- 固定するバージョン: {Round N-1 or Round N}
- 理由: {判定理由}
- 禁止する変更: {具体的に何をしてはいけないか}
```

4. revert 後に `just test` + `just fmt` + `just lint` を実行して整合性を確認する
5. directive が作成された場合、次ラウンド（Round 3）に進む

#### 振動未検出時

振動が検出されなければ directive は作成せず、ループを終了する。Round 3 は実行されない。

### simplify（レビュー修正完了後に1回）

レビュー指摘対応で追加・修正されたコードの冗長性を整理する。`/simplify` を Skill ツールで実行する。

動作を変えずに、重複コード・不要な中間変数・過剰な抽象化などを検出・修正する。修正後に `just test`、`just fmt`、`just lint` を実行して問題がないことを確認する。

### コメント整理（simplify 後に1回）

`agents/comment-cleanup.md` を Read で読み込んだ Haiku モデルのサブエージェント（`model: haiku`）を起動し、`// 動作:` / `// エラー:` 形式の動作コメントをドキュメントコメントに変換する。変換ルールが明確に定義されたルールベース変換タスクのため Haiku で十分な精度が得られる。
