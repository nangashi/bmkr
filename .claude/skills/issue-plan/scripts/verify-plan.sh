#!/usr/bin/env bash
# verify-plan.sh — 計画の事実検証と受け入れ条件カバレッジチェック
#
# Usage: bash verify-plan.sh <plan.md> [issue-body.md]
#
# plan.md: 計画ファイル（必須）
# issue-body.md: Issue 本文ファイル（任意。指定時は受け入れ条件カバレッジもチェック）

set -euo pipefail

PLAN_FILE="${1:?Usage: verify-plan.sh <plan.md> [issue-body.md]}"
ISSUE_FILE="${2:-}"

ERRORS=0
WARNINGS=0

echo "=== 計画の事実検証 ==="
echo ""

# --- 1. 変更ファイル一覧の存在チェック ---
echo "## ファイル存在チェック"
echo ""

# 変更ファイル一覧のテーブルから「変更」操作のファイルパスを抽出
# パターン: | `path` | 変更 | ...  or  | path | 変更 | ...
grep -E '^\|.*\|.*変更.*\|' "$PLAN_FILE" | while IFS='|' read -r _ file_col _ rest; do
    # バッククォートとスペースを除去
    filepath=$(echo "$file_col" | sed 's/`//g' | xargs)
    [ -z "$filepath" ] && continue
    # ヘッダ行をスキップ
    echo "$filepath" | grep -qE '^-+$|^ファイル$' && continue

    if [ -f "$filepath" ]; then
        echo "  OK: $filepath"
    else
        echo "  NG: $filepath (存在しない)"
        ERRORS=$((ERRORS + 1))
    fi
done

echo ""

# --- 2. ADR 参照チェック ---
echo "## ADR 参照チェック"
echo ""

# ADR-NNNN or ADR-0014 形式の参照を抽出
grep -oE 'ADR-[0-9]+' "$PLAN_FILE" | sort -u | while read -r adr_ref; do
    adr_num=$(echo "$adr_ref" | sed 's/ADR-//')
    # ゼロ埋め4桁に正規化
    adr_num_padded=$(printf "%04d" "$adr_num")
    matches=$(find docs/adr/ -name "${adr_num_padded}-*" 2>/dev/null | head -1)
    if [ -n "$matches" ]; then
        echo "  OK: $adr_ref -> $(basename "$matches")"
    else
        echo "  NG: $adr_ref (docs/adr/ に存在しない)"
        ERRORS=$((ERRORS + 1))
    fi
done

echo ""

# --- 3. 受け入れ条件カバレッジ（Issue 本文が指定された場合） ---
if [ -n "$ISSUE_FILE" ] && [ -f "$ISSUE_FILE" ]; then
    echo "## 受け入れ条件カバレッジ"
    echo ""

    # Issue 本文から受け入れ条件を抽出（- [ ] で始まる行）
    ac_count=0
    covered=0
    uncovered_items=""

    while IFS= read -r line; do
        ac_count=$((ac_count + 1))
        # 受け入れ条件のテキストから主要キーワードを抽出（最初の数単語）
        ac_text=$(echo "$line" | sed 's/^- \[.\] //')
        # 計画の受け入れ条件対応表にこのテキストの一部が含まれるかチェック
        if grep -qF "$ac_text" "$PLAN_FILE" 2>/dev/null; then
            echo "  OK: $ac_text"
            covered=$((covered + 1))
        else
            # 部分一致も試す（最初の20文字）
            short_text=$(echo "$ac_text" | cut -c1-20)
            if grep -qF "$short_text" "$PLAN_FILE" 2>/dev/null; then
                echo "  OK: $ac_text (部分一致)"
                covered=$((covered + 1))
            else
                echo "  NG: $ac_text (計画でカバーされていない可能性)"
                uncovered_items="$uncovered_items\n  - $ac_text"
                WARNINGS=$((WARNINGS + 1))
            fi
        fi
    done < <(grep -E '^\- \[.\]' "$ISSUE_FILE")

    echo ""
    echo "  カバレッジ: $covered / $ac_count"

    if [ -n "$uncovered_items" ]; then
        echo ""
        echo "  未カバーの可能性がある条件:"
        echo -e "$uncovered_items"
    fi

    echo ""
fi

# --- サマリ ---
echo "=== 検証結果サマリ ==="
echo "  エラー: $ERRORS"
echo "  警告: $WARNINGS"

if [ "$ERRORS" -gt 0 ]; then
    echo ""
    echo "エラーがあります。計画の該当箇所を確認してください。"
    exit 1
fi

exit 0
