#!/usr/bin/env bash
set -euo pipefail

ADR_DIR="docs/adr"
DEFAULT_THRESHOLD=90
TODAY=$(date +%s)

has_stale=0

printf "| %-6s | %-60s | %-14s | %-8s | %-14s |\n" "ADR" "Title" "last-validated" "Days" "Status"
printf "|--------|--------------------------------------------------------------|----------------|----------|----------------|\n"

for file in "$ADR_DIR"/*.md; do
  [ -f "$file" ] || continue

  # Parse frontmatter (between first and second ---)
  in_frontmatter=0
  status=""
  last_validated=""
  adr_date=""
  review_interval=""
  title=""

  while IFS= read -r line; do
    if [[ "$line" == "---" ]]; then
      if [[ $in_frontmatter -eq 1 ]]; then
        break
      fi
      in_frontmatter=1
      continue
    fi
    if [[ $in_frontmatter -eq 1 ]]; then
      case "$line" in
        status:*)    status=$(echo "$line" | sed 's/^status:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}/\1/') ;;
        date:*)      adr_date=$(echo "$line" | sed 's/^date:[[:space:]]*//') ;;
        last-validated:*) last_validated=$(echo "$line" | sed 's/^last-validated:[[:space:]]*//') ;;
        review-interval:*) review_interval=$(echo "$line" | sed 's/^review-interval:[[:space:]]*//') ;;
      esac
    fi
  done < "$file"

  # Extract title from first H1
  title=$(grep -m1 '^# ' "$file" | sed 's/^# //')

  # Skip withdrawn ADRs
  [[ "$status" == "withdrawn" ]] && continue

  # Use last-validated, fallback to date
  check_date="${last_validated:-$adr_date}"
  if [[ -z "$check_date" ]]; then
    continue
  fi

  # Calculate days elapsed
  check_epoch=$(date -d "$check_date" +%s 2>/dev/null) || continue
  days_elapsed=$(( (TODAY - check_epoch) / 86400 ))

  # Determine threshold
  threshold="${review_interval:-$DEFAULT_THRESHOLD}"

  # Determine status
  if [[ $days_elapsed -gt $threshold ]]; then
    mark="!! STALE"
    has_stale=1
  else
    mark="ok"
  fi

  # ADR number from filename
  adr_num=$(basename "$file" | grep -oE '^[0-9]+')

  # Truncate title if too long
  if [[ ${#title} -gt 60 ]]; then
    title="${title:0:57}..."
  fi

  printf "| %-6s | %-60s | %-14s | %6s d | %-14s |\n" "$adr_num" "$title" "$check_date" "$days_elapsed" "$mark"
done

echo ""
if [[ $has_stale -eq 1 ]]; then
  echo "FAIL: stale ADRs found (threshold: ${DEFAULT_THRESHOLD} days)"
  echo "Run '/adr re-evaluate NNNN' to re-evaluate."
  exit 1
else
  echo "OK: all ADRs are fresh."
fi
