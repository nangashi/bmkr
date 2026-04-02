# L1: 機械的変更フロー

新規ロジックなし、型/I/F 定義不要の変更。I/F 削除、命名変更、import 置換など。

```
Phase 1: 直接編集 + 検証
Phase 2: 品質ゲート
Phase 3: ゴール整合性 + PR 作成
Phase 4: 振り返り
```

---

## Phase 1: 直接編集 + 検証

メインエージェント（Opus）が直接コード編集する。

1. proto/SQL 変更があれば生成コマンドを実行（`just generate`）
2. 計画に従いコードを編集
3. 検証:
   - `just test`
   - `just fmt`
   - `just lint`
4. 変更ファイル一覧と実行コマンドを `edit-log.md` に記録（Phase 4 の入力用）

検証に失敗した場合は修正して再検証する（最大 3 回）。3 回失敗で ABORT。

---

## Phase 2: 品質ゲート

`references/quality-gate.md` に従う。

---

## Phase 3: ゴール整合性 + PR 作成

`references/goal-alignment.md` に従い、受け入れ条件トレーサビリティとスコープ逸脱をチェックする。

チェック通過後、Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/pr-creation.md` を読み込み、コミット・プッシュ・PR 作成を自律実行する。

---

## Phase 4: 振り返り

`references/retrospective.md` に従う。L1 固有の観点:

- 変更が本当に L1 で適切だったか（L2 にすべきだったケースの検出）
- edit-log.md の検証失敗回数
