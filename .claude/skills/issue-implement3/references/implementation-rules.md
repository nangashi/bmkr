# 実装ルール（全レベル共通）

全レベルの実装フェーズで遵守する共通ルール。改善判定方式・ラウンド上限・実装主体は各 flow ファイルで定義する。

---

## best_commit 管理

### 初期化

実装ループ開始前に成果物をコミットする:

```bash
git add -A && git commit -m "checkpoint: before implementation loop (issue #{number})"
```

このコミットが best_commit の初期値（safe point）となる。

### 更新ルール

- **改善時**: checkpoint コミットして best_commit を HEAD に更新する
  ```bash
  git add -A && git commit -m "checkpoint: implementation loop attempt {attempt} (issue #{number})"
  ```
- **停滞/悪化時**: 次ラウンド冒頭で best_commit に復元する（`git reset --hard {best_commit}`）
- **中間成果物**（`.output/` や node_modules 等）は復元時に削除しない

### failure_log

オーケストレータが各 attempt の結果から構成する。次ラウンドのプロンプトに注入して同じ失敗の繰り返しを防ぐ。

```
## 過去の失敗（同じアプローチを避けること）
- Attempt 1: DB接続の初期化をinit()で行ったがテスト時にモック差し込み不可。
  失敗テスト: TestHandleProductList_DBError
  → init()ではなくコンストラクタインジェクションを使うこと。
- Attempt 2: ...
```

情報源の優先順位:
1. サブエージェントの自由記述出力（試したアプローチと失敗理由が直接得られる）
2. テスト出力（どのテストがどう失敗したか）
3. `best_commit` との差分（何を変更したか）

---

## Sonnet サブエージェントへの共通指示

L2 の実装（Opus 直接）では不要。L3 の cluster 実装や品質ゲートの修正で Sonnet サブエージェントを起動する際に以下を指示する:

- `git diff main` で実装対象の差分を確認すること
- 計画の方針・スコープ外
- `docs/guides/` 配下のガイドに従うこと（特に `go-logging.md` と `implementation-anti-patterns.md`）
- テスト実行コマンド（`just test`）と期待される状態（全テスト Green）
- 実装後に `just fmt` と `just lint` を実行して問題があれば修正すること
- 公開 API（ハンドラのルーティング、レスポンス構造等）は変更しないこと。非公開 I/F の修正が必要と判断した場合は、理由と変更範囲を出力して停止すること
- 指示された変更だけを実装すること。明示的に求められていない防御的コード（フォールバック、後方互換性シム、panic recovery、feature detection、graceful degradation）を追加しないこと
- **停止条件**: テストに複数回失敗した、またはテスト自体の誤りが疑われる場合は、失敗理由・試したアプローチ・テスト誤りを疑う根拠を出力して停止すること

---

## テスト保護ルール

### テスト期待値の変更禁止

テストの期待値を実装に合わせて変更することは禁止する。テストが失敗した場合、実装を修正して Green にする。

**例外**: 仕様変更に伴うテスト更新は正当。ただし以下を記録すること:
- invariant statement: 何が変わり、何が変わらないか
- 変更理由: 計画のどの受け入れ条件に基づく仕様変更か

### 既存テストの削除禁止

既存テストを削除して Green にすることは禁止する。

**例外**: 以下のケースのみ許可（invariant statement を記録）:
- 仕様廃止: 計画に明記された RPC/機能の削除に伴うテスト削除
- DI 方式変更: インターフェース抽出に伴うモック再設計で、旧テストを新テストに置き換える

### カバレッジの維持

テストカバレッジを下げる変更は禁止する。テスト削除が正当な場合でも、代替テストで同等のカバレッジを維持する。

---

## wip コメントの扱い

- `// wip:` プレフィックスの動作コメントは、テスト導出の入力として使用する
- 実装完了後（全テスト Green 後）に削除する
- 品質ゲートの Step 4（コメント整理）で doc コメントに変換される

---

## 生成コードの扱い

- proto 変更 → `buf generate`（または `just generate`）
- SQL 変更 → `sqlc generate`（または `just generate`）
- templ 変更 → `templ generate`（または `just generate`）
- 生成後に `go mod tidy` を実行
- 生成コードのテストは書かない（`.claude/rules/go-test.md` ルール 6）

---

## ABORT 時の振る舞い

実装が不可能と判断した場合:

1. best_commit（最も多くのテストが通過した時点）の状態で停止する
2. 以下をユーザーに報告する:
   - 現在の best_commit ハッシュ
   - failure_log サマリ（残存する失敗テストとその原因推定）
   - ABORT 理由
   - 手動継続のための指針
3. 振り返りフェーズに進む（中断振り返り）
