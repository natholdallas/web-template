#!/usr/bin/env bash
# Lint + format gate for the Go backend.
#
# Why a script instead of just `golangci-lint run`: golangci-lint 2.14's built-in
# gofumpt formatter is unreliable (see .golangci.yml), so formatting is checked
# with the standalone binaries, which are deterministic and agree with `gopls`.
#
# Usage:  ./hack/lint.sh          (check)
#         ./hack/lint.sh --fix    (apply formatting fixes first)
set -uo pipefail
cd "$(dirname "$(readlink -f "$0")")/.." || exit 1

GOBIN="${GOBIN:-$HOME/.local/share/go/bin}"
fail=0

step() { printf '\n\033[1;36m== %s\033[0m\n' "$1"; }
ok()   { printf '\033[0;32m%s\033[0m\n' "$1"; }
bad()  { printf '\033[0;31m%s\033[0m\n' "$1"; fail=1; }

if [ "${1:-}" = "--fix" ]; then
  step "Formatting (gofumpt + gci)"
  "$GOBIN/gofumpt" -w .
  "$GOBIN/gci" write -s standard -s default -s 'Prefix(webtplmst)' .
fi

step "Formatting check"
unformatted=$("$GOBIN/gofumpt" -l . 2>/dev/null | grep -v '^$' || true)
if [ -n "$unformatted" ]; then
  bad "gofumpt wants to rewrite:"
  echo "$unformatted"
  echo "run: ./hack/lint.sh --fix"
else
  ok "gofumpt: clean"
fi

unsorted=$("$GOBIN/gci" diff -s standard -s default -s 'Prefix(webtplmst)' . 2>/dev/null || true)
if [ -n "$unsorted" ]; then
  bad "gci wants to reorder imports (run ./hack/lint.sh --fix)"
else
  ok "gci: clean"
fi

step "go build"
go build ./... || bad "build failed"

step "go vet"
go vet ./... || bad "vet failed"

step "staticcheck"
"$GOBIN/staticcheck" ./... || bad "staticcheck failed"

step "golangci-lint"
"$GOBIN/golangci-lint" run ./... || bad "golangci-lint failed"

echo
if [ "$fail" -ne 0 ]; then
  bad "LINT FAILED"
  exit 1
fi
ok "ALL CHECKS PASSED"
