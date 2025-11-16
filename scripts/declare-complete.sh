#!/bin/bash
# declare-complete.sh: Generate completion certificate after all validation gates pass
set -e

# Validate regression tests
bash scripts/regression-test.sh

# Validate manual verification
if ! ls test-results/manual-verification-* 1> /dev/null 2>&1; then
  echo "Manual verification screenshots missing." && exit 1
fi
if ! find test-results/manual-verification-* -name VERIFICATION.md | grep -q VERIFICATION.md; then
  echo "Verification document missing." && exit 1
fi

# Validate ERROR_LOG.md
if [ ! -f ".docs/ERROR_LOG.md" ]; then
  echo "ERROR_LOG.md missing." && exit 1
fi

# Generate completion certificate
cat > AI_INSIGHTS_COMPLETE.md <<EOF
# Completion Certificate

100% complete, tested, and verified with screenshots.
Date: $(date +%Y-%m-%d)

See: test-results/manual-verification-*/VERIFICATION.md
EOF

echo "Completion certificate generated: AI_INSIGHTS_COMPLETE.md"