# L3: 構造的変更フロー

公開 I/F 変更、不変条件の定義が必要、DB スキーマ変更を含む変更。I/F 変更、構造リファクタ、マイグレーションなど。L2+L3 混合ケースもこのフローで実行する。

```
Phase 1: 契約設計
Phase 2: 契約レビュー
Phase 3: テスト戦略 + Red 生成
Phase 4: テストレビュー
Phase 5: 失敗分類
Phase 6: 実装
Phase 7: 品質ゲート
Phase 8: ゴール整合性 + PR 作成
Phase 9: 振り返り
```

---

## Phase 1: 契約設計

Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/contract-design.md` を読み込み、`references/contract-design.md` に従って以下を定義する:

1. proto/SQL 変更があれば生成コマンドを先に実行（`just generate`）
2. contract.md を作成:
   - **Public Interface**: 公開 I/F のシグネチャ
   - **Error Contract**: エラー種別とステータスコード
   - **Invariants**: 変更しない公開 I/F、既存レスポンス形状、ルーティング、主要副作用、スコープ外制約（何を壊さないか）
   - **Generated Code Preconditions**: sqlc 生成型の前提、nullable フィールドの扱い
   - **Acceptance Mapping**: 受け入れ条件と I/F の対応表
3. コード内に `// wip:` 動作コメントを併記（実装時の即時参照用）
4. スタブ定義 + ビルド確認
5. 設計判断を `contract-decisions.md` に記録

### L2 からの昇格時

L2 の interface-design-decisions.md と failure_log を入力として受け取る。白紙から設計せず、L2 の成果物を基に契約として格上げする。既存の wip コメント・型定義は再利用する。

### L2+L3 混合ケース

L2 部分は contract.md の Invariants に「変更しないもの」として記載し、テスト戦略で簡易判定とする。

---

## Phase 2: 契約レビュー

`references/design-review.md` に従い、Codex レビュー + 採用判定 + 修正を行う。

- レビュー観点: 型設計 / エラー契約 / テスト駆動性（perspectives は自己完結型）
- **L3 固有**: perspectives に contract.md の関連セクションをインラインで埋め込む。Codex に渡す際、perspectives + インライン化された契約情報で外部ファイル参照を不要にする
- CRITICAL/IMPORTANT の指摘のみ修正（Opus が修正）

---

## Phase 3: テスト戦略 + Red 生成

### Step 1: テスト戦略

`references/test-strategy.md` に従い、変更対象ごとに以下を判定する:

- **UT 必須**: 新規ロジック、分岐、エラーハンドリング
- **統合テストで十分**: オーケストレーション中心のハンドラ、モックコストが高いケース
- **新規テスト不要**: 機械的変更、既存テストでカバー済み

`test-strategy.md` に記録。`docs/guides/testing-strategy.md` を必ず参照する。

### Step 2: Red 生成

Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/red-generation.md` を読み込み、contract.md + test-strategy.md からテストを導出する。failure cluster に分けやすい粒度で生成する。

`just test` で Red 確認。

---

## Phase 4: テストレビュー

v2 の Red レビューとテスト戦略レビューを 1 回に統合。`references/test-review.md` に従い、Codex レビュー + 採用判定 + 修正を行う。

Codex に渡す perspectives（テスト駆動性観点）に 2 セクションを設ける:

1. **テスト戦略の妥当性**: test-strategy.md の分類判断をレビュー
2. **テストの品質**: テストコードの契約追従・粒度・実装非依存をレビュー

入力最適化: test-strategy.md は **サマリーのみ** をインライン化し、テストコード本体はファイルパスで渡す。

---

## Phase 5: 失敗分類

`references/failure-clustering.md` に従い、Red 実行結果を failure cluster に分類する。

担当: Sonnet（`agents/failure-classifier.md` で独立分類）→ Opus（処理順決定）

---

## Phase 6: 実装

`references/implementation-rules.md` の共通ルールに従う。

### cluster 単位実装

`references/implementation-loop.md` に従う。

- cluster ごとに Sonnet サブエージェントで最小修正（`agents/cluster-implement.md`）
- 検証: 狭いテスト → 広いテスト（`just test`）
- 改善判定: failure 集合差分ベース（どの failure が解消され、どの failure が新規発生したか）
- 改善した cluster のみ checkpoint

### 実装ループ上限

全 cluster の合計実装試行回数: **最大 10 回**（想定 cluster 3〜4 個、cluster あたり平均 2〜3 回）。上限到達時は現在の状態で Phase 7 に進む。scope.md に上限到達を記録する。

### Escalation

`references/escalation.md` に従う。以下の条件で発火する:

- 同一 cluster の連続失敗（2 回）
- compile_gap / tool_behavior_gap の継続
- Sonnet サブエージェントが停止条件を報告（不自然な複雑さ）

発火時の判断:
- **契約に問題あり** → Phase 1 に戻り契約修正、Phase 2（契約レビュー）も再実行
- **テストに問題あり** → Phase 3 に戻りテスト修正、Phase 4（テストレビュー）も再実行
- **実装の問題** → Sonnet で最終 1 回だけ別方針を試す
- **解決不能** → 現在の best state で停止し、ユーザーに報告

### 公開 API 変更の制約

- 公開 API の変更は原則禁止
- 公開 API を変える場合は「契約不足」または「ADR 違反是正」の場合に限る

### 証拠要件

修正時に必ず `.output/issue-implement3/{issue_number}/escalation.md` に記録する:

1. failing evidence
2. invariant statement
3. minimal patch scope

### wip クリーンアップ

全 cluster 完了後、`// wip:` コメントを削除する。

---

## Phase 7: 品質ゲート

`references/quality-gate.md` に従う。

---

## Phase 8: ゴール整合性 + PR 作成

`references/goal-alignment.md` に従い、受け入れ条件トレーサビリティとスコープ逸脱をチェックする。

### L3 固有: Invariants 検証（必須）

contract.md の Invariants セクションに列挙された不変条件が全て維持されていることを確認する:
- `just test` の通過（Phase 7 の修正による退行を検出するため再実行）
- Sonnet によるコードレビュー（Invariants リストと現在のコードを照合）

未対応の受け入れ条件、または Invariants 違反がある場合は Phase 6（実装）に戻る（最大 2 回。2 回目以降は品質ゲートをスキップし Phase 8 のみ再実行）。

チェック通過後、Issue 番号を渡して Sonnet サブエージェントを起動する。サブエージェントは `agents/pr-creation.md` を読み込み、コミット・プッシュ・PR 作成を自律実行する。

---

## Phase 9: 振り返り

`references/retrospective.md` に従う。L3 固有の観点:

- 契約レビュー・テストレビューの空振り率
- Escalation 発火回数と原因の傾向
- 実装ループ上限 10 回の到達有無（到達した場合は上限の妥当性を評価）
- L2 からの昇格で到達した場合、元の分類基準の改善提案
- failure cluster の分類精度（test_bug や tool_behavior_gap が多い場合は契約設計の改善を提案）
