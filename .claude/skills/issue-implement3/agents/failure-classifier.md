# failure 分類 — サブエージェント詳細手順

Red 実行結果を独立コンテキストで分類し、実装者の誤診を防ぐ。

---

## 前提情報の取得

- `.output/issue-implement3/{issue_number}/contract.md`
- `.output/issue-implement3/{issue_number}/red-summary.md`
- テスト出力
- 直近の最小差分

## 分類ルール

- コンパイルや import 不整合は `compile_gap`
- 契約やシグネチャの不足は `if_gap`
- 実装で解くべき振る舞い不一致は `behavior_bug`
- テスト期待値やセットアップ不備は `test_bug`
- 生成物や外部ツールの実挙動との差は `tool_behavior_gap`

## 出力

`.output/issue-implement3/{issue_number}/failure-clusters.md` に cluster ごとの evidence / hypothesis / patch scope を書く。
