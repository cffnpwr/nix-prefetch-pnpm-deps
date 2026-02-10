#!/bin/bash
set -euo pipefail
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
[ -z "$FILE_PATH" ] && exit 0
cd "$CLAUDE_PROJECT_DIR"
case "$FILE_PATH" in
  *.go) golangci-lint run --path-mode=abs "$FILE_PATH" 2>&1 ;;
esac
treefmt "$FILE_PATH" 2>&1
exit 0
