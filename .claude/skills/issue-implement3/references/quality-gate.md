# 品質ゲート

全レベルで実行する共通の品質ゲート。

---

## Step 1: ベースライン

以下を順に実行し、全て通過することを確認する:

1. `just test`（全テスト通過）
2. `just fmt`（フォーマット）
3. `just lint`（リント）

失敗した場合は修正して再実行する。

---

## Step 2: Codex レビュー（最大 3 ラウンド）

`references/codex-review-prompt.md` のテンプレートに以下の 3 観点を指定して Codex レビューを実行する:

- `perspectives/silent-failure.md`（サイレントフェイル）
- `perspectives/project-compliance.md`（プロジェクト準拠）
- `perspectives/security.md`（セキュリティ）

perspectives は自己完結型（contract.md 非依存）。全レベルで同一の perspectives を使用する。

### 採用判定

`agents/review-filter.md` を Read で読み込んだ Sonnet モデルの採用判定サブエージェント（`model: sonnet`）を起動する。

渡すパラメータ:
- `issue_number`
- `review_output_path`（`.output/issue-implement3/{issue_number}/review-final.md`）
- `output_path`（`.output/issue-implement3/{issue_number}/review-final-filtered.md`）

4 軸で判定する:

- **正確性**: 指摘はコードの事実に基づいているか
- **重大度**: CRITICAL / IMPORTANT / LOW
- **スコープ**: Issue の受け入れ条件と計画の範囲内か
- **費用対効果**: 修正コストに見合う効果があるか

CRITICAL / IMPORTANT かつスコープ内の指摘のみ採用する。

### 修正

採用された指摘を Sonnet サブエージェントが修正する（指摘に沿った bounded な修正のため Sonnet で十分）。

### レビュー・修正ループの制御

```
# 品質ゲート開始前のチェックポイント
git add -A && git commit -m "checkpoint: before quality gate (issue #{number})"

for round in 1..3:
  # Round 3 は振動検出時のみ実行
  if round == 3 and oscillation_directives が空:
    break

  Codex 観点付きレビュー → 採用判定（directive があれば filter に渡す）
  採用指摘なし → ループ終了

  Sonnet 修正（fix + test + fmt + lint、directive があればプロンプトに注入）
  git add -A && git commit -m "checkpoint: quality gate round {round} (issue #{number})"

  # Round 2 以降: 振動検出
  if round >= 2:
    振動検出を実行（後述）

simplify（冗長コード整理）
コメント整理（simplify 後に1回）
```

### Codex 観点付きレビュー

`references/codex-review-prompt.md` のテンプレートに以下のパラメータを埋めてプロンプトを構築し、`timeout 600 codex exec --full-auto` に stdin で渡す:

- `{diff_command}`: `git diff main`
- `{perspective_files}`: `silent-failure.md`, `project-compliance.md`, `security.md`
- `{output_path}`: `.output/issue-implement3/{issue_number}/review-final.md`

### 修正（Sonnet サブエージェント）

採用された指摘を Sonnet サブエージェント（`model: sonnet`）に渡し、修正を実行させる。サブエージェントには以下を指示する:

- `git diff main` で現在の差分を確認し、採用された指摘に対応してコードを修正すること
- 修正後に `just test`、`just fmt`、`just lint` を実行して問題がないことを確認すること
- `.output/issue-implement3/{issue_number}/oscillation-directives.md` が存在する場合、その内容を「振動回避指示」としてプロンプトに含める。directive に記載された変更は再度行わないこと

### 振動検出（Round 2 以降、オーケストレータが実行）

各ラウンドの修正後、振動を検出する。

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
2. 保持するバージョンを明示する
   - Round N-1 を採用する場合: 今ラウンドの変更だけを戻す
   - Round N を採用する場合: revert せず現在の状態を保持する
3. revert が必要な場合もファイル全体は戻さず、振動している hunk / ブロックだけを対象にする
4. directive を `.output/issue-implement3/{issue_number}/oscillation-directives.md` に追記する:

```markdown
## Directive {番号}
- ファイル: {file}
- 固定するバージョン: {Round N-1 or Round N}
- 理由: {判定理由}
- 禁止する変更: {具体的に何をしてはいけないか}
```

5. revert または保持判断の反映後に `just test` + `just fmt` + `just lint` を実行して整合性を確認する
6. directive が作成された場合、次ラウンド（Round 3）に進む

#### 振動未検出時

振動が検出されなければ directive は作成せず、ループを終了する。Round 3 は実行されない。

---

## Step 3: simplify

`/simplify` スキルを実行する。

---

## Step 4: コメント整理

Haiku サブエージェント（`agents/comment-cleanup.md`）を起動し、`// wip:` から変換された動作コメントを doc コメントに整理する。
