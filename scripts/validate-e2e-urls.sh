#!/usr/bin/env bash
set -euo pipefail

# Validate E2E tests to ensure relative paths are used (no hardcoded http://localhost:3000 unless docs or env scripts)

echo "Looking for occurrences of 'http://localhost:3000' in tests/e2e (code files only)..."
ERR=0
while IFS= read -r -d '' file; do
  if grep -nH "http://localhost:3000" "${file}" | grep -vE "\.md|README|PLAYWRIGHT_SETUP|setup-test-env" >/dev/null; then
    echo "Found hard-coded host in: ${file}"
    grep -n "http://localhost:3000" "${file}" | sed -n '1,20p'
    ERR=1
  fi
done < <(find tests/e2e -type f -name "*.ts" -o -name "*.js" -print0)

if [ ${ERR} -eq 1 ]; then
  echo "ERROR: Found one or more hard-coded http://localhost:3000 in tests/e2e. Please convert to relative paths or dynamic origin via PLAYWRIGHT_BASE_URL."
  exit 2
fi

echo "No hard-coded http://localhost:3000 found in tests/e2e (code files)."
exit 0
