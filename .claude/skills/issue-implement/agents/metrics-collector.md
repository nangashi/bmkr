# 生成精度メトリクス計測 — サブエージェント詳細手順

AI生成コードの品質を定量的に把握するため、メトリクスを計測し Issue コメントに記録する。

---

## 入力パラメータ（親エージェントから渡される）

- `issue_number`: Issue 番号
- `pass_count`: Phase 4 最終時点のテスト通過数
- `total_count`: Phase 4 最終時点のテスト総数
- `attempt_count`: Phase 4 の実装試行回数（1-4）
- `escalated`: ESCALATE が発生したか（true / false）

---

## 計測手順

### 1. AI生成行数の算出

Phase 3 完了コミット（`checkpoint: phase3 complete` メッセージのコミット）から最新コミットまでの差分を算出する:

```bash
# Phase 3 完了コミットを特定
phase3_commit=$(git log --oneline --grep="checkpoint: phase3 complete" -1 --format="%H")

# 差分の行数を取得（対象: *.go, *.ts, *.tsx）
git diff --stat "$phase3_commit" HEAD -- '*.go' '*.ts' '*.tsx'
```

追加行数 + 削除行数の合計を「AI生成行数」とする。

Phase 3 完了コミットが見つからない場合（Phase 3 がスキップされた場合）は `git diff --stat main HEAD` を使用する。

### 2. テスト通過率の算出

入力パラメータから算出する:

```
テスト通過率 = pass_count / total_count * 100
```

### 3. Issue コメントへの記録

以下のフォーマットで Issue コメントを投稿する:

```bash
gh issue comment {issue_number} --body "$(cat <<'EOF'
<!-- issue-implement:metrics -->
## 生成精度メトリクス

| 指標 | 値 |
|------|-----|
| AI生成行数（追加+削除） | {N} 行 |
| テスト通過率（Phase 4 最終） | {M}% |
| 実装試行回数 | {K} |
| ESCALATE | あり / なし |

計測時刻: {ISO 8601}
EOF
)"
```

`{N}`, `{M}`, `{K}` は計測値で置換する。ISO 8601 は `date -u +"%Y-%m-%dT%H:%M:%SZ"` で取得する。
