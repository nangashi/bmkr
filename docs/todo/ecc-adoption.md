# ECC (everything-claude-code) から取り入れる仕組み

参照元: https://github.com/affaan-m/everything-claude-code

## 優先度：高

### 1. PreCompact Hook — コンテキスト圧縮前の状態保存

コンテキストウィンドウが上限に達して自動圧縮される直前に PreCompact フックが発火し、現在の作業状態をファイルに保存する。

**bmkr への効果**: issue-implement のような長いワークフロー（I/F設計→TDD→実装→レビュー→PR）の途中で自動圧縮が起きると、調査結果や設計判断の文脈が消える。PreCompact フックで「今どのフェーズか」「何を決めたか」を `.output/` に保存すれば、圧縮後も作業を継続できる。

**実装内容**:
- `settings.json` に PreCompact フックを追加
- 小さなシェルスクリプトで現在のタスク状態・直近の決定事項をファイルに書き出す

**実装コスト**: 低

---

### 2. Session Start Hook — セッション開始時のコンテキスト復元

SessionStart フックで前回セッションの要約やプロジェクトの現在状態をロードし、新セッションでもすぐに文脈を持った状態で作業開始できる。

**bmkr への効果**: 現状、新セッションでは毎回 CLAUDE.md を読む程度。SessionStart で「未完了の Issue/PR」「前回の作業サマリ」「直近の ADR 変更」等を自動表示すれば、立ち上がりが速くなる。issue-plan で中断した場合の再開にも役立つ。

**実装内容**:
- `settings.json` に SessionStart フックを追加
- `git log --oneline -5`、`gh issue list`、`gh pr list` 等をまとめて表示するスクリプト

**実装コスト**: 低〜中

---

### 3. Strategic Compaction — 戦略的圧縮の提案

ツール呼び出し回数をカウントし、閾値（50回）に達したら「論理的な区切りで `/compact` を実行してください」と提案する。自動圧縮（95%到達時）はタスク途中で発動するため品質が劣化するが、フェーズの切れ目で手動圧縮すれば重要な文脈を意図的に残せる。

**bmkr への効果**: issue-implement は「I/F設計→テスト導出→実装→レビュー」と明確なフェーズがある。フェーズ完了時に圧縮を促せば、次フェーズに必要な情報だけを残した効率的なコンテキストで作業できる。

**実装内容**:
- PostToolUse フックで Edit/Write 呼び出し回数をカウンターファイルにインクリメント
- 閾値到達時にメッセージを出力

**実装コスト**: 低

---

### 4. Rules — 常時ロードされるルール

`docs/guides/` のような参照ドキュメントとは別に、`.claude/rules/` に常にロードされる短いルールを配置。コーディングスタイル・テスト要件・セキュリティ等を強制する。

**bmkr への効果**: 現在の `docs/guides/` は14件あるが、スキルから参照する設計であり、通常の会話では読み込まれない。頻出するアンチパターン（over-abstraction、cargo-cult-error-handling 等）を短い `.claude/rules/` ルールに抽出すれば、スキル外の通常作業でも常時適用される。

**実装内容**:
- 既存ガイドから要点を抽出し `.claude/rules/` にルールファイルとして配置
- 言語共通ルールと Go/TypeScript 固有ルールに分離

**実装コスト**: 低

---

## 優先度：中

### 5. Continuous Learning（簡易版）— セッションからのパターン抽出

全ツール呼び出しを記録 → バックグラウンドで分析 → 繰り返しパターンを検出 → スキルやルールに昇格。

**bmkr への効果**: 学習用プロジェクトなので、「LLM がどんなミスを繰り返すか」を自動検出してガイドに反映できると有用。現在の `docs/guides/` パターンカタログの拡充を半自動化できる。

**実装方針**: フルシステム（observer daemon + Python CLI + Haiku API）は過剰。簡易版として以下を段階導入:
1. SessionEnd フックでセッションの失敗パターンを `.output/learnings/` にログ
2. 定期的に手動レビューしてガイド化（project-doctor に組み込み可能）

**実装コスト**: 簡易版は中、フル版は高

---

## 取り入れる必要が低いもの

| 仕組み | 理由 |
|--------|------|
| Plankton（高度な品質ゲート） | PostToolUse で oxfmt/oxlint/golangci-lint を既に実行しており十分 |
| AgentShield（セキュリティスキャン） | 学習用プロジェクトで設定の脆弱性スキャンは優先度低 |
| Multi-agent orchestration | issue-implement が既にマルチエージェント構成を実現済み |
| Session aliases | issue-plan が Issue コメントに中間結果を残す設計で代替済み |
| MCP Discipline（ツール数制限） | 現在 context7 + deepwiki の2つだけで問題なし |
