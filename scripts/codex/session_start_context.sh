#!/usr/bin/env bash
set -euo pipefail

cat >/dev/null

cat <<'EOF'
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Read AGENTS.md, docs/repo-structure.md, relevant ADRs, threat-model docs, and the active plan artifact before non-trivial changes. HDIP is identity infrastructure: security, privacy, standards, and traceability outrank speed."
  }
}
EOF
