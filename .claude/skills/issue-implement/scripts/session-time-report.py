#!/usr/bin/env python3
"""セッションログからフェーズマーカーの所要時間を算出して表示する。

Usage:
    python3 .claude/skills/issue-implement/scripts/session-time-report.py [SESSION_LOG_PATH]

SESSION_LOG_PATH を省略した場合、CWD に対応するプロジェクトディレクトリから
最新のセッションログを自動検出する。

セッションログの turn_duration エントリ（ターンごとのアクティブ時間）を使い、
ユーザーの確認待ち時間を除いたアクティブ時間を算出する。
ターンが複数フェーズにまたがる場合は interval overlap で按分する。
turn_duration が存在しないセッションではウォールクロック時間のみ表示する。
"""

import json
import os
import re
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path


def find_session_log() -> str | None:
    """CWD に対応する最新のセッションログを探す。"""
    cwd = re.sub(r"[^a-zA-Z0-9]", "-", os.getcwd())
    project_dir = Path.home() / ".claude" / "projects" / cwd
    if not project_dir.is_dir():
        return None
    logs = sorted(project_dir.glob("*.jsonl"), key=lambda p: p.stat().st_mtime, reverse=True)
    return str(logs[0]) if logs else None


def extract_phase_label(text_line: str) -> str | None:
    """テキスト行からフェーズマーカーを抽出する。

    対応フォーマット:
      - [Phase 0: 初期化]        ← 推奨（SKILL.md で規定）
      - ## Phase 0: 初期化       ← 自然な Markdown 見出し
    """
    stripped = text_line.strip()
    if stripped.startswith("[Phase ") and stripped.endswith("]"):
        return stripped
    if stripped.startswith("## Phase "):
        label = stripped.lstrip("#").strip()
        return f"[{label}]"
    return None


def parse_ts(ts: str) -> datetime:
    return datetime.fromisoformat(ts.replace("Z", "+00:00"))


def fmt_duration(seconds: float) -> str:
    m, s = divmod(int(seconds), 60)
    if m >= 60:
        h, m = divmod(m, 60)
        return f"{h}h {m:02d}m {s:02d}s"
    return f"{m}m {s:02d}s"


def parse_log(log_path: str) -> tuple[
    list[tuple[str, datetime]],
    list[tuple[datetime, datetime]],
    datetime | None,
]:
    """セッションログをパースする。

    Returns:
        markers: [(label, timestamp), ...]
        turns: [(start_dt, end_dt), ...]  ← turn_duration から算出
        last_dt: 最後のエントリのタイムスタンプ
    """
    markers: list[tuple[str, datetime]] = []
    turns: list[tuple[datetime, datetime]] = []
    seen_labels: set[str] = set()
    last_dt: datetime | None = None

    with open(log_path) as f:
        for line in f:
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
                continue

            ts = obj.get("timestamp", "")
            if ts:
                last_dt = parse_ts(ts)
            entry_type = obj.get("type", "")

            # turn_duration の収集
            if entry_type == "system" and obj.get("subtype") == "turn_duration":
                duration_ms = obj.get("durationMs", 0)
                if ts and duration_ms > 0:
                    end_dt = parse_ts(ts)
                    start_dt = end_dt - timedelta(milliseconds=duration_ms)
                    turns.append((start_dt, end_dt))
                continue

            # フェーズマーカーの収集
            if entry_type != "assistant" or not ts:
                continue
            msg = obj.get("message", {})
            if not isinstance(msg, dict):
                continue
            content = msg.get("content", "")
            texts: list[str] = []
            if isinstance(content, list):
                for block in content:
                    if isinstance(block, dict) and block.get("type") == "text":
                        texts.append(block.get("text", ""))
            elif isinstance(content, str):
                texts.append(content)
            for text in texts:
                for text_line in text.splitlines():
                    label = extract_phase_label(text_line)
                    if label and label not in seen_labels:
                        seen_labels.add(label)
                        markers.append((label, parse_ts(ts)))

    return markers, turns, last_dt


def compute_overlap(
    turn_start: datetime, turn_end: datetime,
    phase_start: datetime, phase_end: datetime,
) -> float:
    """ターンとフェーズの重なり秒数を返す。"""
    overlap_start = max(turn_start, phase_start)
    overlap_end = min(turn_end, phase_end)
    return max(0.0, (overlap_end - overlap_start).total_seconds())


def main() -> None:
    if len(sys.argv) > 1:
        log_path = sys.argv[1]
    else:
        log_path = find_session_log()
    if not log_path or not os.path.isfile(log_path):
        print("セッションログが見つかりませんでした。")
        sys.exit(1)

    markers, turns, last_dt = parse_log(log_path)
    if not markers:
        print("フェーズマーカーが見つかりませんでした。")
        sys.exit(0)

    end_time = last_dt or datetime.now(timezone.utc)
    far_future = datetime(2099, 1, 1, tzinfo=timezone.utc)
    has_turns = len(turns) > 0

    print()
    print("## 所要時間")
    print()
    if has_turns:
        print("| Phase | 経過時間 | アクティブ時間 |")
        print("|-------|---------|--------------|")
    else:
        print("| Phase | 経過時間 |")
        print("|-------|---------|")

    total_wall = 0.0
    total_active = 0.0

    for i, (label, phase_start) in enumerate(markers):
        phase_end = markers[i + 1][1] if i + 1 < len(markers) else end_time
        wall = (phase_end - phase_start).total_seconds()
        total_wall += wall

        if has_turns:
            active = sum(
                compute_overlap(t_start, t_end, phase_start, phase_end)
                for t_start, t_end in turns
            )
            total_active += active
            print(f"| {label} | {fmt_duration(wall)} | {fmt_duration(active)} |")
        else:
            print(f"| {label} | {fmt_duration(wall)} |")

    if has_turns:
        print(f"| **合計** | **{fmt_duration(total_wall)}** | **{fmt_duration(total_active)}** |")
    else:
        print(f"| **合計** | **{fmt_duration(total_wall)}** |")


if __name__ == "__main__":
    main()
