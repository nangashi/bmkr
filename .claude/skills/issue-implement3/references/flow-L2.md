# L2: 局所的変更フロー

新規ロジックあり、公開 I/F 変更なし or 追加のみ。新規 RPC 追加、バグ修正、小規模機能追加など。

```
Phase 1: I/F 設計
Phase 2: 設計レビュー（新規 I/F 定義時のみ）
Phase 3: テスト + Red 生成
Phase 4: 実装
Phase 5: 品質ゲート
Phase 6: ゴール整合性 + PR 作成
Phase 7: 振り返り
```

---

## Phase 1: I/F 設計

Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/interface-design.md` を読み込み、以下を実行する:

1. proto/SQL 変更があれば生成コマンドを先に実行（`just generate`）
2. 型・インターフェース・シグネチャを定義
3. `// wip:` 動作コメントを記述（テスト導出の入力になる）
4. ビルド確認（`just build` or コンパイル）
5. 設計判断を `interface-design-decisions.md` に記録

---

## Phase 2: 設計レビュー

### スキップ判断

Phase 1 で **新規の public 型・public インターフェース・public 関数シグネチャ** が定義されなかった場合はスキップする。バグ修正で既存関数内のロジックを追加するだけのケースが該当する。

### 実施する場合

`references/design-review.md` に従い、Codex レビュー + 採用判定 + 修正を行う。

- レビュー観点: 型設計 / エラー契約 / テスト駆動性（perspectives は自己完結型）
- CRITICAL/IMPORTANT の指摘のみ修正

---

## Phase 3: テスト + Red 生成

### スキップ判断

Phase 1 で `// wip:` 動作コメントが記述されなかった場合はスキップする。動作コメントがない = テスト対象の新規ロジックが存在しない。

### 実施する場合

1. wip コメントから UT 対象を特定し `.output/issue-implement3/{issue_number}/test-strategy.md` に簡易記録（UT 対象の一覧のみ。独立レビューは行わない）。L2→L3 昇格時は、この簡易記録を L3 Phase 3 のテスト戦略策定の入力として使用する
2. Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/test-derivation.md` を読み込み、wip 動作コメントからテストを導出する
3. `just test` で Red 確認（コンパイル通過 + テスト失敗が正しい状態。コンパイルエラーは I/F 定義の問題なので修正する）

---

## Phase 4: 実装

`references/implementation-rules.md` の共通ルールに従う。

### 実装ループ（最大 3 ラウンド）

1. メインエージェント（Opus）が実装する
2. `just test` で検証
3. 改善判定: pass_count ベース。前ラウンドより改善していれば続行

### 3 ラウンド到達時: L3 昇格判定

3 ラウンドで Green にならなかった場合、以下を判断する:

1. **失敗の原因が I/F 設計にあるか?**（テスト期待値と実装の不整合、責務分割の問題）
   - → **L3 に昇格**。scope.md を更新し、`references/flow-L3.md` を読み込んで Phase 1 から再開する。L2 の interface-design-decisions.md と failure_log を L3 の契約設計の入力として渡す。既存テストコードは L3 の Phase 3 で再評価する（全面書き直しではない）
2. **テスト/フィクスチャの誤りか?**
   - → テスト/I/F を修正して実装ループを再開（最大 2 回追加。修正が許されるもの: フィクスチャの誤り、動作コメント解釈ミス、テストセットアップ不備。許されないもの: テスト期待値を実装に合わせる変更、カバレッジを下げる変更）
3. **上記いずれでもない**
   - → ABORT。best_commit の状態で停止し、failure_log サマリと ABORT 理由をユーザーに報告。Phase 7（振り返り）に進む

### wip クリーンアップ

全テスト Green 後、`// wip:` コメントを削除する。

---

## Phase 5: 品質ゲート

`references/quality-gate.md` に従う。

---

## Phase 6: ゴール整合性 + PR 作成

`references/goal-alignment.md` に従い、受け入れ条件トレーサビリティとスコープ逸脱をチェックする（Invariants 検証は L3 のみ。L2 では contract.md が存在しないためスキップされる）。

未対応の受け入れ条件がある場合は Phase 4（実装）に戻る（最大 2 回。2 回目以降は品質ゲートをスキップし Phase 6 のみ再実行）。

チェック通過後、Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/pr-creation.md` を読み込み、コミット・プッシュ・PR 作成を自律実行する。

---

## Phase 7: 振り返り

`references/retrospective.md` に従う。L2 固有の観点:

- 設計レビューの空振り率（指摘ゼロだったか）
- L3 昇格が発生したか（発生した場合、分類基準の改善提案）
- テスト導出で過不足があったか
