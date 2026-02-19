#!/usr/bin/env bash
set -euo pipefail

# Task 03 URL-input manual checks using live Pulumi Registry URLs.
# This validates:
# 1) URL fetch path is exercised (non-fetch preflight error on a real schema URL),
# 2) URL non-200 errors are clear and parseable.
#
# REQUIRES_INTERNET=1

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

work_dir="$SCRIPT_DIR/work.url-live"
mkdir -p "$work_dir"

eks_stdout="$work_dir/eks.stdout.txt"
eks_stderr="$work_dir/eks.stderr.txt"
eks_missing_stdout="$work_dir/eks-missing.stdout.txt"
eks_missing_stderr="$work_dir/eks-missing.stderr.txt"

if run_capture "$eks_stdout" "$eks_stderr" \
  --in "https://www.pulumi.com/registry/packages/eks/schema.json" \
  --component "eks:index:Cluster"; then
  eks_code=0
else
  eks_code=$?
fi

# Current expected behavior for this real schema:
# - URL fetch succeeds
# - translation preflight rejects unsupported #/resources refs
assert_eq "$eks_code" "2" "EKS URL exit code"
assert_empty_file "$eks_stdout"
assert_file_contains "$eks_stderr" 'error: schema preflight path'
assert_file_contains "$eks_stderr" 'unsupported ref "#/resources/'
assert_file_contains "$eks_stderr" 'this tool currently supports only local type refs'
if grep -q -- 'fetch input schema URL:' "$eks_stderr"; then
  echo "unexpected URL fetch error for EKS schema URL" >&2
  cat "$eks_stderr" >&2
  exit 1
fi
if [[ "$(awk 'END{print NR}' "$eks_stderr")" != "1" ]]; then
  echo "stderr should be single-line parseable output (EKS URL)" >&2
  cat "$eks_stderr" >&2
  exit 1
fi

if run_capture "$eks_missing_stdout" "$eks_missing_stderr" \
  --in "https://www.pulumi.com/registry/packages/eks/does-not-exist.json" \
  --component "eks:index:Cluster"; then
  missing_code=0
else
  missing_code=$?
fi

assert_eq "$missing_code" "2" "EKS missing URL exit code"
assert_empty_file "$eks_missing_stdout"
assert_file_contains "$eks_missing_stderr" 'error: fetch input schema URL: unexpected status 404 for https://www.pulumi.com/registry/packages/eks/does-not-exist.json'
if [[ "$(awk 'END{print NR}' "$eks_missing_stderr")" != "1" ]]; then
  echo "stderr should be single-line parseable output (missing URL)" >&2
  cat "$eks_missing_stderr" >&2
  exit 1
fi

cat "$eks_stderr"
cat "$eks_missing_stderr"
